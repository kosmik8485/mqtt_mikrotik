#!/usr/bin/env perl
use strict;
use warnings;

use AnyEvent;
use AnyEvent::MQTT;
use Mikrotik::API;
use JSON;

my %Conf = (
  router => {
    host     => '192.168.88.1',
    username => $ENV{'ros_username'},
    password => $ENV{'ros_password'},
    use_ssl  => 0,
    autoconnect => 1,
  },
  mqtt   => {
    host => 'srv.rpi',
    port => 1883,
    user_name => $ENV{'mqtt_user'},
    password => $ENV{'mqtt_pass'},
  },
);

my $mqtt = AnyEvent::MQTT->new( $Conf->{mqtt} );
my $ros = Mikrotik::API->new( $Conf->{router} );

my $cv =AnyEvent->condvar;
my $stop = 0;

my $stop_signal = AnyEvent->signal(
  signal => 'INT',
  cb     => sub { $stop = 1; },
);

my $check_timer = AnyEvent->timer(
  after => 1,
  interval => 1,
  cb => sub { $cv->send if $stop; }
);

my $work_timer = AnyEvent->timer(
  adter => 1,
  interval => 10,
  cb => sub {
    my $data = JSON->new->utf8->encode( getWiFiUsers );
    $mqtt->publish(
      topic => "router/home",
      message => $data,
    );
  },
);

$cv->recv;
exit;

sub getWiFiUsers {
  my ( $ret_print, @users ) = $ros->query("/interface/wireless/registration-table/print");
  return \@users;
}
