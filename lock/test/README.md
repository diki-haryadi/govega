## Usage

```
go build
./test // using local lock, shared lock
./test -u redis://localhost:6379 // using redis
./test -m skip // using local lock non-shared lock
./test -t 4 // set thread to 3 , default is 2
./test -i 20 // set iteration, default is 10
```