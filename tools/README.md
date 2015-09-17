# GoHeroes2 Tools

## Extractor

The extractor can extract game assets from AGG archives

### Benchmark

Here's a comparison extracting the the `HEROES2.AGG` and `HEROES2X.AGG` archives.

 * `HEROES2.AGG`: 1434 assets
 * `HEROES2X.AGG`: 52 assets

free heroes2:

    $ time src/tools/extractor data/HEROES2.AGG aggex4
    src/tools/extractor data/HEROES2.AGG aggex4  0.17s user 0.47s system 43% cpu 1.475 total
    src/tools/extractor data/HEROES2X.AGG aggex5  0.04s user 0.03s system 50% cpu 0.146 total

GoHeroes2

    $ go build tools/extractor.go
    $ time ./extractor data/HEROES2.AGG output
    ./extractor data/HEROES2.AGG output  0.09s user 0.35s system 31% cpu 1.366 total
    ./extractor data/HEROES2X.AGG output  0.01s user 0.02s system 54% cpu 0.051 total

At present the only downside is the Go binary is ~2.3mb vs 190kb for the C++ one. Yikes!
