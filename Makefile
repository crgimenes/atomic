@all:
	go build -o atomic -v ./cmd/atomic
	go build -o atomicdb -v ./cmd/atomicdb
	go build -o ipterm -v ./cmd/ipterm
	go build -o iptclient -v ./cmd/iptclient
	go build -o keyboard -v ./cmd/keyboard

install:
	cp  ./atomic ~/bin/
	cp  ./atomicdb ~/bin/
	cp  ./ipterm ~/bin/
	cp  ./iptclient ~/bin/
	cp  ./keyboard ~/bin/

clean:
	rm -f atomic
	rm -f atomicdb
	rm -f ipterm
	rm -f iptclient
	rm -f keyboard
