# Introduction

This is a barebones alternative [AppImage](https://github.com/probonopd/AppImageKit) [type 2](https://github.com/AppImage/AppImageSpec/blob/master/draft.md#type-2-image-format) runtime (image loader) and image maker.  The image is an application and a ZIP archive.

# Usage

```sh
go get github.com/orivej/static-appimage/...
make-static-appimage APPDIR DESTINATION
```

`APPDIR` must already contain an `AppRun`.
