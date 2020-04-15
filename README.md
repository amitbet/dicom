
# A production ready parser in golang:
This started out as a fork of [suyashkumar's](https://github.com/suyashkumar/dicom) excellent repo, 
but since I had a lot of things to add, some i wanted to remove (I didn't need the actual pixel data while parsing) 
and since golang has the path burnt into the package names, I ended up moving away from the original repo
My purpose is to create something that can be used in a production ready dicom-web server.

The intended usecase:
* Indexing the metadata in a DB for qido queries 
* Indexing each frame location within files
* Storing the files on disk / S3 bucket
* Serving DicomWeb (WadoRS), reading each frame from disk as required

My Modifications:
- [x] Fixed some bugs with image data stored inside sequences
- [x] Added DicomWeb compatible Json formatting for DataSet objects
- [x] Added file offset and size data to parsed frames (for seeking within stored dicoms)
- [x] Added support for odd field data
- [x] General refactoring of parser options and parsing code
- [x] Tested on additional dicom files (various modality types and image encodings)
- [ ] TODO: Add support for deflate transfer syntax

## Usage
To use this in your golang project, import `github.com/amitbet/dicom` and then you can use `dicom.Parser` for your parsing needs:
```go 
p, err := dicom.NewParserFromFile("myfile.dcm", nil)
opts := dicom.ParseOptions{DropPixelData: true}

element := p.ParseNext(nil) // parse and return the next dicom element
// or
dataset, err := p.Parse(opts) // parse whole dicom
```
More details about the package can be found in the [godoc](https://godoc.org/github.com/amitbet/dicom)

## CLI Tool
A CLI tool that uses this package to parse imagery and metadata out of DICOMs is provided in the `dicomutil` package. All dicom tags present are printed to STDOUT by default. 

```
dicomutil --extract-images-stream myfile.dcm
```

## Acknowledgements
* Suyashkumar's [dicom project](https://github.com/suyashkumar/dicom)
* Original [go-dicom](https://github.com/gillesdemey/go-dicom)
* Grailbio [go-dicom](https://github.com/grailbio/go-dicom) -- commits from their fork were applied to ours
* GradientHealth for supporting work I did on this while there [gradienthealth/dicom](https://github.com/gradienthealth/dicom)
* Innolitics [DICOM browser](https://dicom.innolitics.com/ciods)
* [DICOM Specification](http://dicom.nema.org/medical/dicom/current/output/pdf/part05.pdf)
* <div>Icons made by <a href="https://www.freepik.com/?__hstc=57440181.48e262e7f01bcb2b41259e2e5a8103b3.1557697512782.1557697512782.1557697512782.1&__hssc=57440181.4.1557697512783&__hsfp=2768524783" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" 			    title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" 			    title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a></div>
