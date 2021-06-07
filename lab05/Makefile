.PHONY: all

all: build clean

build: primary secondary

primary: MatmulPrimary.java
	hadoop com.sun.tools.javac.Main MatmulPrimary.java
	jar cf primary.jar MatmulPrimary*.class

secondary: MatmulSecondary.java
	hadoop com.sun.tools.javac.Main MatmulSecondary.java
	jar cf secondary.jar MatmulSecondary*.class

clean:
	rm -f *.class