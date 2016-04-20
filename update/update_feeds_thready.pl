#!/usr/bin/perl
use strict;
use DBI;
use threadeach;
use Config::Std;
use List::Util qw/shuffle/;
`killthe update_feeds.pl`;

my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");


use Data::Dumper;
print Dumper \%config;
my $dbh = get_dbh(%config);

my $sql=qq{select id  from ttrss_feeds where 1};

my $sth=$dbh->prepare($sql);
$sth->execute();
my @all_ids;
while(my ($id)=$sth->fetchrow_array)
{
	push(@all_ids,$id);
}

threadeach my $id(shuffle(@all_ids))
{
	system("perl update_feeds.pl feed_id=$id");
}
print scalar(@all_ids)." feeds updated\n";



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