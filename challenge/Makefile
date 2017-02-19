include ./Makefile.defs

TARGET=./bin/challenge-executable

all: $(TARGET)

run:
	./bin/challenge-executable

clean:
	go clean
	rm -f $(TARGET)

SOURCES := $(shell find ./daemon ./common ./pkg . -name '*.go')

$(TARGET): $(SOURCES)
	go build -o $(TARGET) ./main.go

# install: all
# 	$(INSTALL) -m 0755 $(TARGET) $(DESTDIR)$(BINDIR)
# 	$(INSTALL) -m 0755 -d $(DESTDIR)$(CONFDIR)/bash_completion.d
# 	$(INSTALL) -m 0644 -T ../contrib/autocomplete/bash_autocomplete $(DESTDIR)$(CONFDIR)/bash_completion.d/cilium
