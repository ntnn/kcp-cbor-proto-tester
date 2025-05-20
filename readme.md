# Benchmark

Run a kcp instance:

```bash
kcp start
```

Run the benchmark once KCP is ready with one core left for KCP to
actually answer requests:

```bash
go test -bench=. -benchmem -parallel $(( $(nproc) - 1 ))
```

`BenchmarkWithClientCM` creates, reads and deletes configmaps, which should
be pretty equivalent in performance as the bulk of it is the data
portion which should be the same across all formats.

`BenchmarkWithClientRBAC` creates, reads and deletes a cluster role, which
should be more representative of the actual performance difference as
this is more structured data.

## Results

Results are from an Apple M4 Pro with 48GB of RAM with kcp on the same
machine.

### Default

| Benchmark | Loops | ns / operation | Bytes / op | allocations/op |
|---|---|---|---|---|
| BenchmarkWithClientCM/application/json-12 |           30 |         534325301 ns/op |           70527 B/op |        767 allocs/op |
| BenchmarkWithClientCM/application/yaml-12 |           36 |         545367865 ns/op |           67575 B/op |        742 allocs/op |
| BenchmarkWithClientCM/application/vnd.kubernetes.protobuf-12 |                37 |         546694412 ns/op |           70382 B/op |        740 allocs/op |
| BenchmarkWithClientCM/application/cbor-12 |                                   33 |         539856298 ns/op |           67882 B/op |        750 allocs/op |
| BenchmarkWithClientRBAC/application/json-12 |                                 46 |         557056567 ns/op |           63813 B/op |        710 allocs/op |
| BenchmarkWithClientRBAC/application/yaml-12 |                                 43 |         554051771 ns/op |           64740 B/op |        717 allocs/op |
| BenchmarkWithClientRBAC/application/vnd.kubernetes.protobuf-12 |       33 |         539861898 ns/op |           68191 B/op |        742 allocs/op |
| BenchmarkWithClientRBAC/application/cbor-12 |                                       51 |         561236383 ns/op |           62617 B/op |        700 allocs/op |
| BenchmarkSerialization/json-12 |          1000000000 |               0.0000240 ns/op |               0 B/op |          0 allocs/op |
| BenchmarkSerialization/yaml-12 |          1000000000 |               0.0000837 ns/op |               0 B/op |          0 allocs/op |
| BenchmarkSerialization/protobuf-12 |      1000000000 |               0.0000040 ns/op |               0 B/op |          0 allocs/op |
| BenchmarkSerialization/cbor-12 |          1000000000 |               0.0000169 ns/op |               0 B/op |          0 allocs/op |


### -benchtime=100x

| Benchmark | Loops | ns / operation | Bytes / op | allocations/op |
|---|---|---|---|---|
| BenchmarkWithClientCM/application/json-12 |                100 |         580233178 ns/op |           61095 B/op |        671 allocs/op |
| BenchmarkWithClientCM/application/yaml-12 |                100 |         580294898 ns/op |           60076 B/op |        669 allocs/op |
| BenchmarkWithClientCM/application/vnd.kubernetes.protobuf-12 |                     100 |         580306646 ns/op |           63327 B/op |        671 allocs/op |
| BenchmarkWithClientCM/application/cbor-12 |                                        100 |         580241004 ns/op |           60250 B/op |        669 allocs/op |
| BenchmarkWithClientRBAC/application/json-12 |                                      100 |         580259841 ns/op |           58981 B/op |        662 allocs/op |
| BenchmarkWithClientRBAC/application/yaml-12 |                                      100 |         580248819 ns/op |           59426 B/op |        663 allocs/op |
| BenchmarkWithClientRBAC/application/vnd.kubernetes.protobuf-12 |                   100 |         580231348 ns/op |           60638 B/op |        663 allocs/op |
| BenchmarkWithClientRBAC/application/cbor-12 |                                      100 |         580248774 ns/op |           59373 B/op |        662 allocs/op |
| BenchmarkSerialization/json-12 |               100 |               726.7  ns/op |            74 B/op |          0 allocs/op |
| BenchmarkSerialization/yaml-12 |               100 |              2412  ns/op |            1194 B/op |          9 allocs/op |
| BenchmarkSerialization/protobuf-12 |           100 |                83.75  ns/op |           24 B/op |          0 allocs/op |
| BenchmarkSerialization/cbor-12 |               100 |               449.2  ns/op |            49 B/op |          0 allocs/op |

