#!/usr/bin/env perl
use strict;
use warnings;

use AnyEvent;
use AnyEvent::MQTT;
use API::MikroTik;
use Getopt::Long;
use JSON;

my %options = (
  router_user => $ENV{'ros_username'},
  router_pass => $ENV{'ros_password'},
  mqtt_user   => $ENV{'mqtt_user'},
  mqtt_pass   => $ENV{'mqtt_pass'},
);

GetOptions(
  'ros_user=s'  => \$options{router_user},
  'ros_pass=s'  => \$options{router_pass},
  'mqtt_user=s' => \$options{mqtt_user},
  'mqtt_pass=s' => \$options{mqtt_pass},
) or die($!);

my %Conf = (
  router => {
    host        => '192.168.88.1',
    port	=> 8728,
    user        => $options{router_user},
    password    => $options{router_pass},
    tls         => 0,
    autoconnect => 1,
  },
  mqtt   => {
    host => 'srv.rpi',
    port => 1883,
    user_name => $options{mqtt_user},
    password => $options{mqtt_pass},
  },
);

my $mqtt = AnyEvent::MQTT->new( $Conf{mqtt} );
my $ros = API::MikroTik->new( $Conf{router} );

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
  interval => 5,
  cb => sub {
    my $data = JSON->new->utf8->convert_blessed->encode( @{[ getWiFiUsers() ]} );    
    $mqtt->publish(
      topic => "router/home",
      message => $data,
    );
  },
);

$cv->recv;
exit;

sub getWiFiUsers {
  my $list = $ros->cmd("/interface/wireless/registration-table/print");
  if ( my $err = $ros->error ){    
    print "[ERROR] $err";    
    $cv->send;
  }
  return $list;
}
