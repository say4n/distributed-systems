.PHONY: all

all: build clean

build: primary secondary

primary: MatmulPrimary.java clean
	hadoop com.sun.tools.javac.Main MatmulPrimary.java
	jar cf primary.jar MatmulPrimary*.class

secondary: MatmulSecondary.java clean
	hadoop com.sun.tools.javac.Main MatmulSecondary.java
	jar cf secondary.jar MatmulSecondary*.class

start-hadoop:
	-$$HADOOP_HOME/sbin/start-all.sh

stop-hadoop:
	-$$HADOOP_HOME/sbin/stop-all.sh

exec-primary: primary.jar
	hadoop fs -mkdir -p /matmul/input
	hadoop fs -put -f input/*.mat /matmul/input
	hadoop fs -ls /matmul/input
	hadoop fs -rm -r -f /matmul/intermediate_output
	hadoop jar primary.jar MatmulPrimary /matmul/input /matmul/intermediate_output

clean:
	rm -f *.class