# UI Perf Validation (S2/S3)

| Dataset | Initial Render (ms) | Nav Median (ms) | Search Median (ms) | Sort (ms) | Detail Open (ms) | Discover Action (ms) | Reports Export (ms) | Crashed |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| S2 | 415 | 50 | 9 | 15 | 34 | 497 | 150 | no |
| S3 | 430 | 73 | 12 | 13 | 48 | 1440 | 147 | no |

- Targets:
  - S2: initial <=1000ms, nav median <=150ms, search median <=300ms, sort <=250ms, detail open <=120ms
  - S3: no crash; discover action <=2000ms; reports export <=2000ms