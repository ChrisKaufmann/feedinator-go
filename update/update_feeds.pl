#!/usr/bin/perl
use strict;
use XML::Feed;
use Encode;
use DBI;
use CGI;
use Config::Std;
use Digest::SHA1 qw(sha1 sha1_hex sha1_base64);
use HTML::Entities;
use Data::Dumper;
use LWP::UserAgent;
use HTTP::Date;
use Date::Parse;
use POSIX qw(strftime);
use Cache::Memcached;

my $cgi=new CGI;
my $debug =$cgi->param('debug') || 0;
my $memd = new Cache::Memcached { 'servers' => [ "127.0.0.1:11211" ], 'debug'=> $debug, 'compress_threshold' => 10_000, };
my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");
my $prefix=$config{'Web'}{'environment'};
my $dbh=get_dbh(%config);

my $feed_key="${prefix}FeedList";
my $errors_key="${prefix}Errors";

my %feeds=get_feeds();
my $insert_entry_sth=$dbh->prepare("insert into ttrss_entries (title,guid,link,updated,content,content_hash,feed_id,comments,no_orig_date,date_entered,user_name) values (?,?,?,?,?,?,?,?,?,NOW(),?)") or die("Couldn't prepare insert_entry_sth: $!");
my $err_sql=$dbh->prepare("update ttrss_feeds set error_string=? where id=? limit 1") || die("Couldn't prepare err_sql: $!");


my %error_skips=();
%error_skips = data($errors_key) if hasdata($errors_key);
if($cgi->param('feed_id')){ update_feed(scalar $cgi->param('feed_id'));exit();}
foreach my $id(keys %feeds)
{
	update_feed($id);
}

sub update_feed
{
	my $id=shift || die("No id passed to update_feed");
	return if $error_skips{$id};
	print "Updating $id(".$feeds{$id}{'link'}."): ";
	my $feed; #holds the parsed feed
	my $category_id=$feeds{$id}{'category_id'};
	my $ua=LWP::UserAgent->new;
#	$ua->default_header('If-Modified-Since' => HTTP::Date::time2str(Date::Parse::str2time($feeds{$id}{'last_updated'})));
	my $response=$ua->get($feeds{$id}{'link'});
	my $time = strftime "%Y-%m-%d %H:%M:%S", localtime($response->headers->date);
	my $new_entries=0;
	if($response->code == 304)
	{
		print "\t304 (no update)\n";
		set_last_updated($id,$time);
	}
	eval{
		my $cc=$response->decoded_content;
		$feed = XML::Feed->parse(\$cc)
	};
	if($@ or !$feed)
	{
		print "\tNo data$1 $@\n";
		set_error($id, "No data: $1 $@");
		set_last_updated($id);
		$error_skips{"$id"}=1;
		set($errors_key,\%error_skips);
		return;
	} else {
	    set_error($id);
	}
	my @guids=();
	foreach my $item($feed->entries)
	{
		my $url=&escape_guid($item->link()) || next;
		push(@guids,$url);
	}
	my %existing_entries=get_existing_entries($id,"'".join("','",@guids)."'");
	my @excludes=split(',',$feeds{$id}{'excludes'});
	foreach my $item($feed->entries)
	{
		my $guid	= escape_guid($item->link()) || next;
		if($existing_entries{$guid})
		{
			print ".";
			next;
		}
		$existing_entries{$guid}=1;
		my $title       =&escape($item->title()) || &escape($feed->title());
		$title          =~s/\n//g;
		if(grep {$title=~m/$_/i } @excludes ){print "skipping $title\n";next;}   #skip items that match the exclude array
		my $desc	    = escape($item->content->body);
		my $desc_hash   = sha1_hex(encode_utf8($desc));
		$desc			= '&nbsp;' unless $desc;
		my $updated		= $item->modified || $time;
		my $comments	= '&nbsp';
		$insert_entry_sth->execute($title,$guid,$guid,$updated,$desc,$desc_hash,$id,$comments,'true',$feeds{$id}{'username'})
			|| warn("Couldn't execute insert_entry_sth: $!");
		print "+";
		$new_entries++;
	}
	print "\n";
	$memd->set("${prefix}Feed${id}ExistingHashes",\%existing_entries);
	print("Deleting ${prefix}Feed${id}_unreadentries\n");
	$memd->delete("${prefix}Feed${id}_unreadentries");
    print("Incrementing ${prefix}Feed${id}_UnreadCount, $new_entries\n");
	$memd->incr("${prefix}Feed${id}_UnreadCount", $new_entries);
    print("Deleting ${prefix}Category${category_id}_unreadentries\n");
	$memd->delete("${prefix}Category${category_id}_unreadentries");
    print("Incrementing ${prefix}Category${category_id}_UnreadCount, $new_entries\n");
	$memd->incr("${prefix}Category${category_id}_UnreadCount", $new_entries);
	set_last_updated($id);
}
sub set_last_updated
{
	my $id=shift || die("NO id passed to set_last_updated");
	my $updated_date = shift || strftime "%Y-%m-%d %H:%M:%S", localtime;
	my $sql = "update ttrss_feeds set last_updated=? where id=?";
	my $sth=$dbh->prepare($sql) or die("Coudln't prepare sth for update last_updated: $!");
	$sth->execute($updated_date,$id) or die("Couldn't execute update last_update: $!");
	$feeds{$id}{'last_updated'}=$updated_date;
	set($feed_key,\%feeds);
}
sub get_existing_entries
{
	my $id=shift || return;
	my $new_guids=shift;
	my %existing=();
	my $mcref="${prefix}Feed${id}ExistingHashes";
	return data($mcref) if hasdata($mcref);
	my $and=$new_guids?"AND guid in ($new_guids)":"";
	my $sql="select guid from ttrss_entries where feed_id='$id' $and";
	my $sth=$dbh->prepare($sql);
	$sth->execute();
	while(my $id=($sth->fetchrow_array())[0])
	{
		$existing{$id}=1;
	}
	set($mcref,\%existing);
	return %existing;
}
sub get_feeds
{
	my %rethash=();
	return data($feed_key) if hasdata($feed_key);
	my $sql=qq{select id, feed_url,user_name,title,exclude,exclude_data,last_updated,category_id from ttrss_feeds where 1};
	my $sth=$dbh->prepare($sql) || die("Couldn't prepare get_feeds sql: $!");
	$sth->execute || die("Couldn't execute get_feeds sql ($sql): $!");
	my $cat_sql=qq{select c.exclude from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and f.id=?};
	my $cat_sth=$dbh->prepare($cat_sql) || die ($!);
	while(my ($id,$link,$username,$title,$exclude,$excludedata,$last_updated,$category_id)=$sth->fetchrow_array)
	{
	    if(!$link=~m/^[http|https]:\/\//) {
		    $link='http://'.$link;
	    }
		$rethash{$id}{'link'}=$link;
		$rethash{$id}{'excludes'}=$exclude;
		$rethash{$id}{'last_updated'}=$last_updated;
		$rethash{$id}{'username'}=$username;
		$rethash{$id}{'category_id'}=$category_id;
		$cat_sth->execute($id)
			|| die("Couldn't execute cat_sth: $!\n");
		if($cat_sth->rows != 0)
		{
			my $cat_exclude=($cat_sth->fetchrow_array())[0];
			if($cat_exclude != "")
			{
				$rethash{$id}{'excludes'}="$exclude,$cat_exclude";
			}
		}
	}
	set($feed_key,\%rethash);
	return %rethash;
}
sub set
{
	my $key=shift;
	my $val=shift || die("Not enough params passed to set");
	$memd->set($key,$val);
}
sub hasdata
{
	my $key=shift || die ("No key passed to hasdata");
	print("C");
	my $ref = $memd->get($key);
	my $firstkey;
	try {$firstkey=(keys %{$ref})[0]}
	catch {return 0;}
	return 1 if $firstkey;
	return 0;
}
sub data
{
	my $key=shift || die ("No key passed to data");
	return %{$memd->get($key)};
}
sub get_dbh
{
	my %c=shift || die ("No config passed to get_dbh");
	my $db_user =$config{'DB'}{'user'};
	my $db_pass =$config{'DB'}{'pass'};
	my $db_host =$config{'DB'}{'host'}?$config{'DB'}{'host'}:'localhost';
	my $db_db =$config{'DB'}{'db'};

	my $dbh=DBI->connect("DBI:mysql:$db_db:$db_host","$db_user","$db_pass")
		|| die("Couldn't connect to db: $!");
	return $dbh;
}
sub escape_guid
{
	my $in=shift;
	$in=~s/'/&#39;/g;
	$in=~s/&#8217;/&#39;/g;
	$in=~s/’/&#39;/g;
	$in=~s/’/&#39;/g;
	$in=~s/’/&#39;/g;
	$in=~s/"/&#34;/g;
	$in=~s/\*/&#42;/g;
	$in=~s/\//&#47;/g;
	$in=~s/\\/&#92;/g;
	$in=~s/“/&#34;/g;
	$in=~s/”/&#34;/g;
	$in=~s/–/-/g;
	$in=~s/—/-/g;
	$in=~s/—/-/g;
	return $in;
}
sub escape
{
	my $in=shift;
	return encode_entities($in);
}
sub try (&$) {
   my($try, $catch) = @_;
   eval { $try };
   if ($@) {
      local $_ = $@;
      &$catch;
   }
}
sub set_error() {
    my $id=shift || return;
    my $err_str=shift || "";
    $err_sql->execute($err_str,$id) || die("couldn't execute set_err sql: $!");
    $memd->delete("Feed${id}_");
    return;
}