test:
	go test -v

bench:
	go test -bench . -benchmem \
		-cpuprofile=cpu.out \
		-memprofile=mem.out \
		-memprofilerate=1 \
		.

pprof-cpu:
	go tool pprof \
		-trim_path=/Users/danilov6083/GolandProjects/stepikGoWebServices/3-pprof/\
		-source_path=~/GolandProjects/stepikGoWebServices/3-profiling/ \
		stepikGoWebServices.test \
		mem.out

pprof-mem:
	go tool pprof \
		-trim_path=/Users/danilov6083/GolandProjects/stepikGoWebServices/3-pprof/\
		-source_path=~/GolandProjects/stepikGoWebServices/3-profiling/ \
		stepikGoWebServices.test \
		mem.out

pprof-cpu-web:
	go tool pprof \
		-trim_path=/Users/danilov6083/GolandProjects/stepikGoWebServices/3-pprof/ \
		-source_path=~/GolandProjects/stepikGoWebServices/3-profiling/ \
		-http=:8083 \
		stepikGoWebServices.test \
		cpu.out

pprof-mem-web:
	go tool pprof \
		-trim_path=/Users/danilov6083/GolandProjects/stepikGoWebServices/3-pprof/ \
		-source_path=~/GolandProjects/stepikGoWebServices/3-profiling/ \
		-http=:8083 \
		stepikGoWebServices.test \
		mem.out
