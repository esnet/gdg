# Security Policy

## Supported Versions

### GDG Version 

We'll strive to support the last 2 minor versions.  Example, at the time of writing of this document, the latest GDG version is 0.7.1.

Any issues with 0.6.X and 0.7.X family are supported.  Anything beyond that should still work, but is used at your own risk.  

| GDG Version | Supported          |
| ------- | ------------------ |
| 0.7.1   | :white_check_mark: |
| 0.6.x   | :white_check_mark:             |
| <= 0.5.x   | :x:                |

### Grafana Version

Since GDG is almost entirely dependant on the Grafana API, support is also tied to the Grafana version.  We'll strive to do our best 
to support the last 2 major versions of Grafana.  At the time of the publishing of this document the current version of Grafana is 11.3.

That means we'll do a best effort to ensure that 11.X and 10.X are supported.  If a breaking change is introduced by Grafana we may move along with Grafana or support both 
behaviors till the next major release of Grafana when we'd drop the support for the legacy behavior.  

That being said, it's likely that most of the older versions will keep on working with older versions of Grafana, as long as the Grafana API has not had any breaking changes.


| Grafana Version | Supported   |
| ------- | ------------------  |
| 11.X   | :white_check_mark:   |
| 10.X   | :white_check_mark:   |
| <= 9.0   | :x:                |

## Reporting a Vulnerability

Coming Soon
