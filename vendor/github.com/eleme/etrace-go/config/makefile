NOW=$(shell date '+%Y-%m-%d %H:%M:%S')
bench:
	@go test -bench=. -benchmem|grep mark|awk 'BEGIN {print "## $(NOW)\n\nmethod | iteration | ns/op | B/op | allocs/op \n-----|-----|-----|-----|-----"}  {print $$1 "|" $$2 "|" $$3  "|" $$5 "|" $$7} END {print "\n\n"}'|tee -a benchmark.md

