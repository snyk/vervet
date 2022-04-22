package service

// AnnotationVUScrape is the annotation vervet-underground looks for to determine whether the service should be scraped.
// Kube Services should be annotated with `vervet-underground.snyk.io/scrape: "true"` to enable scraping.
const AnnotationVUScrape = "vervet-underground.snyk.io/scrape"

// AnnotationVUPort specifies the port to scrape from. If omitted, port 80 is used by default.
const AnnotationVUPort = "vervet-underground.snyk.io/port"
