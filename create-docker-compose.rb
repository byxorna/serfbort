#!/usr/bin/env ruby

require 'yaml'

@nagent = ARGV[0].to_i || 2
$stderr.puts "Creating docker-compose.yml for 1 master and #{@nagent} agents"

# the docker-compose.yaml
@compose = {
  "serfnet" => {
    "image" => "google/pause"
  },
  "master" => {
    "build" => ".",
    "command" => "master -name master",
    "net" => "container:serfnet",
    "ports" => ["7373:7373"]
  }
}

#@links = ["master"]
#(0..@nagent).each{|x| @links << "agent#{x}"}

(0...@nagent).each do |nagent|

  name = "agent#{nagent}"
  listenport = 7947+nagent
  @compose[name] = {
    "build"=>".",
    "command" => "-config examples/config.json agent -name #{name} -master localhost:7946 -listen localhost:#{listenport}",
    "net" => "container:serfnet",
    #"links" => @links.select{|x| x!=name}.to_a
  }
end

puts @compose.to_yaml
