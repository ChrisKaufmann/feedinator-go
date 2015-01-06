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

my $cgi=new CGI;
my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");
my $db_user =$config{'DB'}{'user'};
my $db_pass =$config{'DB'}{'pass'};
my $db_host =$config{'DB'}{'host'}?$config{'DB'}{'host'}:'localhost';
my $db_db   =$config{'DB'}{'db'};
my $debug   =$cgi->param('debug') || 0;
my $dbh=DBI->connect("DBI:mysql:$db_db:$db_host","$db_user","$db_pass") 
	|| die("Couldn't connect to db: $!");
my %namehash=();
my %titlehash=();
my %excludehash=();
my %excludedatahash=();
my %updatedhash=();

my ($second, $minute, $hour, $dayOfMonth, $month, 
		$yearOffset, $dayOfWeek, $dayOfYear, $daylightSavings) = localtime();
		$yearOffset = 1900 + $yearOffset;
		$month++;$month='0'.$month if length($month)<2;
		$dayOfMonth='0'.$dayOfMonth if length($month)<2;
		my $today_time="$yearOffset-$month-$dayOfMonth $hour:$minute:$second";

my %feed_list=&get_feedlist();
my $count=0;
my @feed_ids=sort(keys %feed_list);
#&shuffle(\@feed_ids);
foreach my $id(@feed_ids)
{
		$count++;
		&update_feed($id,$feed_list{$id},$namehash{$id});
}
$dbh->do("analyze table ttrss_entries;");
print "$count feeds updated\n" unless $cgi->param('feed_id');
exit();

sub update_feed
{
	my $id=shift			|| return;
	my $source=shift		|| return;
	my $username=shift;
	my $feed;
	my $ua=LWP::UserAgent->new;
	print "id=$id, source=$source, updated_time=$updatedhash{$id}\n";
	$ua->default_header('If-Modified-Since' => HTTP::Date::time2str(Date::Parse::str2time($updatedhash{$id})));
	my $response=$ua->get($source);
	if($response->code == 304)
	{
		print("No update, skipping\n");
		my $updatetime=strftime "%Y-%m-%d %H:%M:%S", localtime($response->headers->date);
		set_last_updated($id,$updatetime);
		return;
	}
	
	eval{
	my $cc=$response->decoded_content;
	$feed = XML::Feed->parse(\$cc)
		or do {
			print XML::Feed->errstr.", feed id=$id\n";
			add_feed_error($id,XML::Feed->errstr);
			return;				
		};
	};  return if $@;
	return unless $feed;
	my @excludes=split(',',$excludehash{$id});
	my @dataexcludes=split(',',$excludedatahash{$id});
	my @guids=();
	foreach my $item( $feed->entries )
	{
		my $url			=&escape_guid($item->link()) || next;
		my $guid		=$url;
		push (@guids,$guid);
	}
	my %existing_entries=&get_existing_entries($id,"'".join("','",@guids)."'");

	print "Title: ".encode_utf8($feed->title())."\n"; 
	print "Date: ". $feed->modified(). "\n";
	print scalar(keys %existing_entries)." existing entries\n";
	print "Num of entries in rss feed: ".scalar($feed->entries)." \n";
	my $sql=qq{insert into ttrss_entries 
			(title,guid,link,updated,content,content_hash,
			feed_id,comments,no_orig_date,date_entered,user_name)
			values(?,?,?,?,?,?,?,?,?,NOW(),?)};
	my $sth=$dbh->prepare($sql);

	foreach my $item( $feed->entries )
	{
		my $url			=&escape_guid($item->link()) || next;
		my $guid		=$url;

		if($existing_entries{$guid})
		{
			print ".";
			next;
		}
		my $title		=&escape($item->title()) || &escape($feed->title());
		$title			=~s/\n//g;
		if(grep {$title=~m/$_/i } @excludes ){print "skipping $title\n";next;}   #skip items that match the exclude array
		my $tmp_content=$item->content;
		my $desc		=&escape($tmp_content->body);
		if(grep {$desc=~m/$_/i } @dataexcludes) {print "skipping $title(data)\n";next;} #same with items in data
		my $desc_hash   =sha1_hex(encode_utf8($desc));
		$desc			='&nbsp;' if !$desc;
		my $updated		=$item->modified || $today_time;
		my $comments	='&nbsp;';
		$sth->execute($title,$guid,$url,$updated,$desc,$desc_hash,
				$id,$comments,'true',$username);
		print "+";
	}
	print "\n";
	set_last_updated($id);
}
sub set_last_updated
{
	my $id=shift || die("NO id passed to set_last_updated");
	my $updated_date = shift || strftime "%Y-%m-%d %H:%M:%S", localtime;
	print "Updated date=$updated_date\n";
	my $sql = "update ttrss_feeds set last_updated=? where id=?";
	my $sth=$dbh->prepare($sql) or die("Coudln't prepare sth for update last_updated: $!");
	$sth->execute($updated_date,$id) or die("Couldn't execute update last_update: $!");
}
sub add_feed_error
{
	my $id=shift || die("No id passed to add_feed_error");
	my $errstr=shift || die("No error string passed to add_feed_error");
	$errstr=~s/['"]/ /g;
	my $sql="update ttrss_feeds set error_string='$errstr' where id=$id";
	`echo "$sql" | mysql -u$db_user -p$db_pass -h$db_host $db_db`;
}
sub set_title
{
		my $id=shift;
		my $title=shift;
		if(!$id or !$title){return;}
		my $sql="update ttrss_feeds set title='$title' where id='$id' limit 1";
		my $sth=$dbh->prepare($sql);
		$sth->execute();
}
sub get_existing_entries
{
		my $id=shift || return;
		my $new_guids=shift;
		my $and=$new_guids?"AND guid in ($new_guids)":"";
		my $sql="select guid from ttrss_entries where feed_id='$id' $and";
		my $sth=$dbh->prepare($sql);
		$sth->execute();
		my %existing=();
		while(my $id=($sth->fetchrow_array())[0])
		{
				$existing{$id}=1;
		}
		return %existing;
}
sub get_feedlist
{
		my %rethash=();
		my $where=$cgi->param('feed_id') ? " id = ".$cgi->param('feed_id')." ": 1;
		my $sql=qq{select id, feed_url,user_name,title,exclude,exclude_data,last_updated from ttrss_feeds where $where};
		my $sth=$dbh->prepare($sql) || die($!);
		$sth->execute() || die("Couldn't execute: $!");
		#get category excludes
		my $cat_sql=qq{select c.exclude from ttrss_feeds as f, ttrss_categories as c where f.category_id=c.id and f.id=?};
		my $cat_sth=$dbh->prepare($cat_sql) || die ($!);
		while(my ($id,$link,$username,$title,$exclude,$excludedata,$last_updated)=$sth->fetchrow_array)
		{
				$link=~s/http:\/\///;
				$link='http://'.$link;
				$rethash{$id}=$link;
				$namehash{$id}=$username;
				$titlehash{$id}=$title;
				$excludehash{$id}=$exclude;
				$excludedatahash{$id}=$excludedata;
				$updatedhash{$id}=$last_updated;
				$cat_sth->execute($id)
					|| die("Couldn't execute cat_sth: $!\n");
				if($cat_sth->rows != 0)
				{
					my $cat_exclude=($cat_sth->fetchrow_array())[0];
					if($cat_exclude != "")
					{
						$excludehash{$id}="$exclude,$cat_exclude";
					}
				}
		}
		return %rethash;
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
#sub shuffle
#{
#		my $array = shift;
#		my $i = @$array;
#		while ( --$i )
#		{
#				my $j = int rand( $i+1 );
#				@$array[$i,$j] = @$array[$j,$i];
#		}
#}
