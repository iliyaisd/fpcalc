## Packages:

`fpcorr`: Go implementation of [this algorithm](https://shivama205.medium.com/audio-signals-comparison-23e431ed2207) for audio comparison using [fpcalc](https://acoustid.org/chromaprint), which implements Chromaprint algorithm. It is assumed that `fpcalc` is present in the PATH. 

Usage example:

```
corr, err := AudioCorrelate("path/to/file1.wav", "path/to/file2.wav")
log.Printf("Correlation: %f", corr)
```
`corr` is a value between 0 and 1 that indicates correlation (the bigger, the more similar two audios are).
