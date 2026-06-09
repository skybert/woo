# woo

A fast, zero config static site generator.

## Usage

```text
$ woo
```

## Features

### DONE
- Copy files in cwd of `.md` to target dir.
- Convert `.md` to `.html`
- Expand template variables

### TODO
- Use article source dir
- Create article target dir (don't write all articles to the category
  directory).
- Generate menu
- Generate tag index
- Generate category index
- Configuration file support
- CLI arguments for most regular use cases, e.g. var expansion with
  `--variable TITLE=Foo --variable SITE_URL=https://example.com`
- HTTP server for site testing, `make serve`

## Install

```text
$ make install
```
