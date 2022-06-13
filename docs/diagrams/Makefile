PUMLS = $(shell find . -type f -name "*.puml")
PNGS  = $(PUMLS:.puml=.png)

.PHONY: all
all: diagrams

.PHONY: diagrams
diagrams: $(PNGS)

plantuml.jar:
	wget -O $@ https://github.com/plantuml/plantuml/releases/download/v1.2022.5/plantuml-1.2022.5.jar --no-check-certificate --quiet

%.png: plantuml.jar %.puml
	@# Using pipelining to preserve file names.
	cat $*.puml | java -jar plantuml.jar -pipe > $*.png
