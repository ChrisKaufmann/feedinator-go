#!/usr/bin/perl
use strict;
use DBI;
use threadeach;
use Config::Std;

my %config={};
read_config("../config", %config) or die("Couldn't read config file: $!");


use Data::Dumper;
print Dumper \%config;

my $db_user	=$config{'DB'}{'user'};
my $db_pass	=$config{'DB'}{'pass'};
my $db_host	=$config{'DB'}{'host'};
my $db_db	=$config{'DB'}{'db'};
my $dbh=DBI->connect("DBI:mysql:$db_db:$db_host","$db_user","$db_pass");

my $sql=qq{select id  from ttrss_feeds where 1};

my $sth=$dbh->prepare($sql);
$sth->execute();
my @all_ids;
while(my ($id)=$sth->fetchrow_array)
{
	push(@all_ids,$id);
}

threadx4 my $id(@all_ids)
{
	system("perl update_feeds.pl feed_id=$id");
}
