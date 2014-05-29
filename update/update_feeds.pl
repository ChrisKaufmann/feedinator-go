#!/usr/bin/perl
use strict;
use XML::Feed;
use Encode;
use DBI;
use CGI;
use Config::Std;
use Digest::SHA1 qw(sha1 sha1_hex sha1_base64);
use HTML::Entities;
print "Updating\n";
my $cgi=new CGI;
my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");
my $db_user =$config{'DB'}{'user'};
my $db_pass =$config{'DB'}{'pass'};
my $db_host =$config{'DB'}{'host'}?$config{'DB'}{'host'}:'localhost';
my $db_db   =$config{'DB'}{'db'};
my $debug   =$cgi->param('debug') || 0;
my $dbh=DBI->connect("DBI:mysql:$db_db:$db_host","$db_user","$db_pass");
my %namehash=();
my %titlehash=();
my %excludehash=();

my ($second, $minute, $hour, $dayOfMonth, $month, 
		$yearOffset, $dayOfWeek, $dayOfYear, $daylightSavings) = localtime();
		$yearOffset = 1900 + $yearOffset;
		$month++;$month='0'.$month if length($month)<2;
		$dayOfMonth='0'.$dayOfMonth if length($month)<2;
		my $today_time="$yearOffset-$month-$dayOfMonth $hour:$minute:$second";

my %feed_list=&get_feedlist();
my $count=0;
my @feed_ids=sort(keys %feed_list);
&shuffle(\@feed_ids);
foreach my $id(@feed_ids)
{
		$count++;
		&update_feed($id,$feed_list{$id},$namehash{$id});
		sleep(2);
}
$dbh->do("analyze table ttrss_entries;");
print "$count feeds updated\n" unless $cgi->param('feed_id');
exit();

sub update_feed
{
	my $id=shift			|| return;
	print "\nID: $id\n";
	my $source=shift		|| return;
	my $username=shift;
	my $feed;
	eval{
	$feed = XML::Feed->parse(URI->new($source))
		or do {
			print XML::Feed->errstr.", feed id=$id\n";
			add_feed_error($id,XML::Feed->errstr);
			return;				
		};
	};  return if $@;
	return unless $feed;
	my @excludes=split(',',$excludehash{$id});
	my %existing_entries=&get_existing_entries($id);

	print "Title: ".encode_utf8($feed->title())."\n"; 
	print "Date: ". $feed->modified(). "\n";
	print scalar(keys %existing_entries)." existing entries\n";
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
		my $desc_hash   =sha1_hex(encode_utf8($desc));
		$desc			='&nbsp;' if !$desc;
		my $updated		=$item->modified || $today_time;
		my $comments	='&nbsp;';
		$sth->execute($title,$guid,$url,$updated,$desc,$desc_hash,
				$id,$comments,'true',$username);
		print "+";
	}
	print "\n";
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
		my $sql="select guid from ttrss_entries where feed_id='$id'";
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
		my $sql=qq{select id, feed_url,user_name,title,exclude from ttrss_feeds where $where};
		my $sth=$dbh->prepare($sql);
		$sth->execute();
		while(my ($id,$link,$username,$title,$exclude)=$sth->fetchrow_array)
		{
				$link=~s/http:\/\///;
				$link='http://'.$link;
				$rethash{$id}=$link;
				$namehash{$id}=$username;
				$titlehash{$id}=$title;
				$excludehash{$id}=$exclude;
		}
		return %rethash;
}

sub escape_guid
{
	my $in=shift;
	$in=~s/'/&#39;/g;
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
	$in=~s/'/&#39;/g;
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
sub shuffle
{
		my $array = shift;
		my $i = @$array;
		while ( --$i )
		{
				my $j = int rand( $i+1 );
				@$array[$i,$j] = @$array[$j,$i];
		}
}
