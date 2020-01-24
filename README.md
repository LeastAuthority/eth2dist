# eth2dist

This repo contains code to determine whether block proposers are sampled uniformly (by their stake).

To run the code, do

```
go test -v
```

Note that `-v` is required to see to relevant output. Depending on your
computer, it may make sense to add `-timeout 30m` or so, because the
computation takes a while and by default Go tests abort after 10 minutes.

## Output

The test first outputs the number of cores detected. It spawns that number of workers for hashing. Each of these workers reports progress every 32 samples. The line contains which sample they are at.

After all tests finish, the distributions will be printed.
- `real` is an array that counts how often each validator has been elected.
- `rand` is an array that counts how often each validator had been selected, if the selection _was uniform_. We sample actual randomness here.
- `mean` is the mean of `real`
- `stddev` is the standard deviation of `real`
- `rmean` is the mean of `rand`
- `rstddev` is the standard deviation of `rand`

## Parameters

the file `election_test.go` contains a `var ( ... )` block with parameters.
On `go test`, all combinations will be executed.

Another thing that isn't tested yet is how the algorithm behaves if the stake is not evenly distributed. I figure the most difficult part about this is capturing how the ideal distribution should look.
