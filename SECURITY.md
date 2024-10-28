# Security Policy

## Supported Versions

### GDG Version 

We'll strive to support the last 2 minor version.  Example, at the time of writing of this document, the latest GDG version is 0.7.1.

Any issues with 0.6.X and 0.7.X family are supported.  Anything beyond that should still work, but is used at your owrk risk.  

| GDG Version | Supported          |
| ------- | ------------------ |
| 0.7.1   | :white_check_mark: |
| 0.6.x   | :white_check_mark:             |
| <= 0.5.x   | :x:                |

### Grafana Version

Since GDG is almost entirely dependant on the grafana API, support is also tied to the Grafana version.  We'll strive to do our best 
to support the last 2 major version of grafana.  At the time of the publishing of this document the current version of grafana is 11.3.

That means we'll do a best effort to ensure that 11.X and 10.X.  If a breaking change is introduced by grafana we may move along with grafana or support both 
behaviors till the next major release of grafana when we'd drop the support for the legacy behavior.  

That being said, it's likely that most of the older versions will keep on working with older versions of grafana, as long as the grafana API has not had any breaking changes.


| Grafana Version | Supported   |
| ------- | ------------------  |
| 11.X   | :white_check_mark:   |
| 10.X   | :white_check_mark:   |
| <= 9.0   | :x:                |

## Reporting a Vulnerability

Coming Soon
