# Testdata Generator Examples

This helper binary generates PNGs with random colors, a diagonal line,
and a centered index label.

## Small icons for a 512x512 atlas

Generate many small sprites (16..64) plus some rectangles for variety.

```bash
rm -rf ./test/512
# Small square-ish icons
go run ./cmd/testdata-generator -m 16 -M 64 -c 25 ./test/512/squares
# Extra rectangles for variety
go run ./cmd/testdata-generator -m 16 -M 64 -n -r 4 -c 50 ./test/512/rects
```

> [!NOTE]  
>
> * Total area is approximate; generate extra and pack as needed.
> * Use -n -r to allow non‑power‑of‑two rectangles.

And generate `imageset` and `edds`

```bash
go run ./cmd/imageset-packer/ pack -fdc -g4 -x1 ./test/512/ /p/ -r bssf
```

## Large 4K atlas split into 4 groups

Each group is generated into its own subfolder. Later you can pack each
group separately or together.

```bash
rm -rf ./test/4k
# Group 1 (small)
go run ./cmd/testdata-generator -m 32 -M 128 -n -r 4 -c 100 test/4k/group_01
# Group 2 (medium)
go run ./cmd/testdata-generator -m 64 -M 256 -n -r 4 -c 50 test/4k/group_02
# Group 3 (large)
go run ./cmd/testdata-generator -m 128 -M 512 -n -r 3 -c 15 test/4k/group_03
# Group 4 (mixed, wider)
go run ./cmd/testdata-generator -m 64 -M 512 -n -r 6 -c 45 test/4k/group_04
# Generate
go run ./cmd/imageset-packer/ pack -fdc -x1 ./test/4k/ /p/
```
