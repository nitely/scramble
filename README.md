# Scramble

Scramble is a data format that encodes byte nibbles by scrambling the bits.
This library supports data stream encoding/decoding to/from Scramble.
This lib uses no dependencies.

## Compile

```
go build scramble.go
```

## Usage

```
$ ./scramble -encode "foo"; echo
$ ./scramble -decode "[124807030]"; echo
$ echo -n "foo" | ./scramble -encode | ./scramble -decode; echo
```

## Test

```
go test ./scrambler
```

## LICENSE

MIT
