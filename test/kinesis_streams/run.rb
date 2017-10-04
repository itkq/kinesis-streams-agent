#!/usr/bin/env ruby

log = "/var/log/kinesis-streams-agent/test.log"

system("touch #{log}")
sleep 1

t = Thread.new do
  sleep(1)
  3.times do
    system("logrotate -f ./logrotate.conf")
    puts "rotated"
    sleep(1)
  end
end

count = 500

count.times do |i|
  File.open(log, "a") do |f|
    ino = File.stat(log).ino
    str = "{\"i\": #{i},\"key\":\"logrotate_test\"}"
	f.puts(str)
    puts("#{ino} -> #{str}")
    sleep 0.01
  end
end

t.join
sleep 1