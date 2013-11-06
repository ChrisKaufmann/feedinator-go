#!/usr/bin/perl;
use strict;
use DBI;
use Config::Std;
use Digest::SHA1 qw(sha1 sha1_hex sha1_base64);

my $cgi=new CGI;
my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");
my $db_user =$config{'DB'}{'user'};
my $db_pass =$config{'DB'}{'pass'};
my $db_host =$config{'DB'}{'host'}?$config{'DB'}{'host'}:'localhost';
my $db_db   =$config{'DB'}{'db'};
my $debug   =$cgi->param('debug') || 0;
my $dbh=DBI->connect("DBI:mysql:$db_db:$db_host","$db_user","$db_pass");



my $sql="select id,expirey from ttrss_feeds where expirey != '' and expirey is not null;";
my $sth=$dbh->prepare($sql);
$sth->execute();
while(my ($id,$expirey)=$sth->fetchrow_array)
{
	if(!$id){next;}
	print "Expiring from feed $id\n";
	my $sql2="delete from ttrss_entries where feed_id='$id' and  date_entered < date_sub(now(),interval $expirey) and marked=0";
	my $sth2=$dbh->prepare($sql2);
	$sth2->execute();
}
