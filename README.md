# Epub-Scraper
[![Go Report Card](https://goreportcard.com/badge/github.com/DaRealFreak/epub-scraper)](https://goreportcard.com/report/github.com/DaRealFreak/epub-scraper)  ![GitHub](https://img.shields.io/github/license/DaRealFreak/epub-scraper)  
Application to scrape novels and convert them into EPUB files based on YAML configuration files.

## Dependencies
- [Calibre](https://calibre-ebook.com/) - cross-platform open-source suite of e-book software.

This dependency is required to fix encoding errors, image compression and to keep latest standards (ebook-polish of it has to be callable)

## Usage
You can simply pass the configuration file you want to process by either dropping them onto the binary
or by passing it in the command line.  
On passing folders to the binary it'll process all available .yaml files from within folder.

```
Usage:
  scraper [file 1] [file 2] ... [flags]
  scraper [command]

Available Commands:
  help        Help about any command
  update      update the application

Flags:
  -h, --help               help for scraper
  -v, --verbosity string   log level (debug, info, warn, error, fatal, panic) (default "info")
      --version            version for scraper
```

## Configuration
To be compatible with most use cases a lot of configurations are possible for the extraction of the e-book source.
Only a few keys are actually required though, so you can generate valid Epub files with a minimal configuration already.

Most minimal configuration with at least 1 chapter would be:
```yaml
general:
  title: [string][required]
  author: [string][required]
chapters:
  - chapter:
      url: [string][required]
      title-content:
        title-selector: [string][required]
      chapter-content:
        content-selector: [string][required]
```

You can also find multiple real usage example configurations in the [examples](examples) folder.

### General
Metadata and Table of Content related information for the generated Epub file.

All available configuration options:
```yaml
general:
  # title of the generated Epub
  title: [string][required]
  # sub title of the generated Epub
  alt-title: [string]
  # author of the generated Epub
  author: [string][required]
  # description of the generated Epub
  description: [string]
  # cover image, can be either a file path or an URL to an image
  cover: [string]
  # language of the generated Epub
  language: [string]
  # link to the original novel
  raw: [string]
  # translators to be mentioned and linked in the Table of Content page
  translators:
    - # displayed name of the translator
      name: [string]
      # URL to link the displayed name to
      url: [string]
```

### Sites
Optional section with the intention to single out the chapter title and content settings by the domain.
Especially useful in case single chapters are getting added in the chapters section.  
Redirects are only configurable in this section. Each redirect configuration is only used if the chapter host matches the site configuration host.
If we get redirected to a different host it'll also use use the site configuration of the new host.

All available configuration options:
```yaml
sites:
  - # host of site
    host: [string][required]
    # possible redirects, it'll try to follow them as deep as possible, else it'll use the next closes URL
    redirects: [list of strings]
    # optional configuration in case the Table of Content has multiple pages
    pagination:
      # should extracted chapters be reversed?
      # allows newest -> oldest navigation to work with unknown amount of pages
      reverse-posts: [boolean]
      # CSS selector to the next page, has to point to an element with an "href" attribute
      next-page-selector: [string]
    # required configurations to extract the chapter titles
    title-content:
      # will add a "Chapter [index+1] - " to the title if true
      add-prefix: [boolean]
      # CSS selector for the title
      title-selector: [string][required]
      # possibility to narrow down title selection by cutting of prefix
      # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
      prefix-selectors: [list of strings]
      # possibility to narrow down title selection by cutting of suffix
      # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
      suffix-selectors: [list of strings]
      # option to clean up the extracted title using regular expressions
      cleanup-regex: [string]
    # required configuration to extract the chapter content
    chapter-content:
      # CSS selector for the chapter content
      content-selector: [string][required]
      # possibility to narrow down title selection by cutting of prefix
      # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
      prefix-selectors: [list of strings]
      # possibility to narrow down title selection by cutting of suffix
      # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
      suffix-selectors: [list of strings]
```

### Chapters
Contains the configuration where to extract chapters from. Either direct links to chapters (chapter) of links to
Table of Content (toc) pages are available.  
One element can't have both chapter and toc at the same time since the order of the chapters would be unknown.
Just append them each as one chapter source.  
If no configuration is set for `title-content` and `chapter-content` it'll use the related site configuration if set.
If chapter source and related site are both configured the chapter source configuration will be preferred over the site configuration.

All available configuration options:
```yaml
chapters:
  # table of content element where we can extract chapters from
  - toc:
      # URL to extract chapters from (and starting point of the navigation if set)
      url: [string][required]
      # CSS selector to the chapter link, has to point to an element with an "href" attribute
      # redirects are possible with the site configuration (for f.e. blog post -> chapter links)
      chapter-selector: [string][required]
      # optional configuration in case the Table of Content has multiple pages
      pagination:
        # should extracted chapters be reversed?
        # allows newest -> oldest navigation to work with unknown amount of pages
        reverse-posts: [boolean]
        # CSS selector to the next page, has to point to an element with an "href" attribute
        next-page-selector: [string]
      # required configurations to extract the chapter titles
      title-content:
        # will add a "Chapter [index+1] - " to the title if true
        add-prefix: [boolean]
        # CSS selector for the title
        title-selector: [string][required]
        # possibility to narrow down title selection by cutting of prefix
        # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
        prefix-selectors: [list of strings]
        # possibility to narrow down title selection by cutting of suffix
        # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
        suffix-selectors: [list of strings]
        # option to clean up the extracted title using regular expressions
        cleanup-regex: [string]
      # required configuration to extract the chapter content
      chapter-content:
        # CSS selector for the chapter content
        content-selector: [string][required]
        # possibility to narrow down title selection by cutting of prefix
        # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
        prefix-selectors: [list of strings]
        # possibility to narrow down title selection by cutting of suffix
        # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
        suffix-selectors: [list of strings]

  # chapter element, direct link to the chapter
  - chapter:
      # direct link to the chapter, redirects possible with the site configuration (for f.e. blog post -> chapter links)
      url: [string][required]
      # required configurations to extract the chapter titles
      title-content:
        # will add a "Chapter [index+1] - " to the title if true
        add-prefix: [boolean]
        # CSS selector for the title
        title-selector: [string][required]
        # possibility to narrow down title selection by cutting of prefix
        # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
        prefix-selectors: [list of strings]
        # possibility to narrow down title selection by cutting of suffix
        # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
        suffix-selectors: [list of strings]
        # option to clean up the extracted title using regular expressions
        cleanup-regex: [string]
      # required configuration to extract the chapter content
      chapter-content:
        # CSS selector for the chapter content
        content-selector: [string][required]
        # possibility to narrow down title selection by cutting of prefix
        # cut off will only occur at first match, so use 2x same prefix if you want to select after the 2nd occurrence
        prefix-selectors: [list of strings]
        # possibility to narrow down title selection by cutting of suffix
        # cut off will only occur after first match, so use 2x same suffix if you want to select before 2nd last occurrence
        suffix-selectors: [list of strings]
```

### Blacklist
You can blacklist URLs of which no chapter data will be extracted. This is useful if you use multiple hosts
to extract chapters which may overlap with each other. The blacklist will also be checked during the redirect checks.

configuration:
```yaml
blacklist: [list of strings]
```


### Assets
The assets section contains information about the assets included in the generated .epub file.
Added assets will be included in every added chapter automatically.
```yaml
assets:
  css:
    # path relative to YAML file to the CSS file used in the generated Epub
    path: [string]
  font:
    # path relative to YAML file to the font file used in the generated Epub
    path: [string]
```

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details
