#!/usr/bin/env ruby

# system("systemctl start kinesis-agent-go")
sleep 1

log = "/var/log/kinesis-agent-go/test.log"

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

# system("systemctl kill -s9 kinesis-agent-go")
# sleep 1
#
# expected = `cat #{log}* | wc -l`.strip
# actual = `wc -l /var/log/kinesis-agent-go/test_output`.strip
#
# if expected == actual
#   exit(0)
# else
#   puts "assert failed"
#   puts "  expected: #{expected}"
#   puts "    actual: #{actual}"
#   exit(1)
# end
