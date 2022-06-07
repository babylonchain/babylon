PUMLS = $(shell find . -type f -name "*.puml" -exec basename {} \;)
PNGS  = $(PUMLS:.puml=.png)

.PHONY: all
all: $(PNGS)

plantuml.jar:
	wget -O $@ https://github.com/plantuml/plantuml/releases/download/v1.2022.5/plantuml-1.2022.5.jar --no-check-certificate --quiet

%.png: plantuml.jar
	java -jar plantuml.jar $*.puml
