.PHONY: all

all: build

build: primary secondary clean

copy:
	ssh bigdataupf 'rm -rf ~/matmul && mkdir -p ~/matmul'
	scp * bigdataupf:~/matmul

primary: MatmulPrimary.java
	hadoop com.sun.tools.javac.Main MatmulPrimary.java
	jar cf primary.jar MatmulPrimary*.class

secondary: MatmulSecondary.java
	hadoop com.sun.tools.javac.Main MatmulSecondary.java
	jar cf secondary.jar MatmulSecondary*.class

start-hadoop:
	-$$HADOOP_HOME/sbin/start-all.sh

stop-hadoop:
	-$$HADOOP_HOME/sbin/stop-all.sh

execute: exec-primary exec-secondary

exec-primary: primary.jar
	hadoop fs -rm -r -f /matmul/input
	hadoop fs -mkdir -p /matmul/input
	hadoop fs -put -f *.mat /matmul/input
	hadoop fs -ls /matmul/input
	hadoop fs -rm -r -f /matmul/intermediate_output
	time hadoop jar primary.jar MatmulPrimary /matmul/input /matmul/intermediate_output

exec-secondary: exec-primary secondary.jar
	hadoop fs -rm -r -f /matmul/output
	time hadoop jar secondary.jar MatmulSecondary /matmul/intermediate_output /matmul/output

clean:
	rm -f *.class