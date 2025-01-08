---
title: "Working with Library Panels"
weight: 32
---

Starting with version 0.4, library panels are going to be supported. It's a bit special and the behavior is somewhat unique.

Rules:
  - Library Panels are immutable.  They cannot be moved to a different folder.  They are linked to one or multiple dashboards.
  - The only way I can see to move a lib element is to unlink the panel, delete the panel and re-create it in a different folder, then re-link it.
  - In theory it's supposed to move with the dashboards but I haven't been able to re-create that behavior.
  - You cannot delete a library element while a dashboard is still using it.


## Import components

will retrieve all the components from Grafana and save to local file system.



```sh
gdg lib download
┌─────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ TYPE    │ FILENAME                                                                                              │
├─────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ library │ testing_data/libraryelements/General/dashboard-makeover-extra-cleaning-duty-assignment-today.json     │
│ library │ testing_data/libraryelements/General/dashboard-makeover-lighting-status.json                          │
│ library │ testing_data/libraryelements/General/dashboard-makeover-side-dish-prep-times-past-7-days.json         │
│ library │ testing_data/libraryelements/General/dashboard-makeover-time-since-we-purchased-these-spices.json     │
│ library │ testing_data/libraryelements/General/extreme-dashboard-makeover-grill.json                            │
│ library │ testing_data/libraryelements/General/extreme-dashboard-makeover-mac-oven.json                         │
│ library │ testing_data/libraryelements/General/extreme-dashboard-makeover-refrigerator-temperature-f.json       │
│ library │ testing_data/libraryelements/General/extreme-dashboard-makeover-room-temperature-f.json               │
│ library │ testing_data/libraryelements/General/extreme-dashboard-makeover-salmon-cooking-times-past-7-days.json │
└─────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Importing Dashboards
Now that we the library components, pulled let's pull the Dashboard.

```sh
gdg dash download
INFO[0002] Importing dashboards for context: 'local'
┌───────────┬───────────────────────────────────────────────────────────────────┐
│ TYPE      │ FILENAME                                                          │
├───────────┼───────────────────────────────────────────────────────────────────┤
│ dashboard │ testing_data/dashboards/General/bandwidth-dashboard.json          │
│ dashboard │ testing_data/dashboards/General/bandwidth-patterns.json           │
│ dashboard │ testing_data/dashboards/Other/dashboard-makeover-challenge.json   │ <== uses library panels
│ dashboard │ testing_data/dashboards/Other/flow-analysis.json                  │
│ dashboard │ testing_data/dashboards/Other/flow-data-for-circuits.json         │
│ dashboard │ testing_data/dashboards/Other/flow-data-for-projects.json         │
│ dashboard │ testing_data/dashboards/Other/flow-data-per-country.json          │
│ dashboard │ testing_data/dashboards/Other/flow-data-per-organization.json     │
│ dashboard │ testing_data/dashboards/Other/flow-information.json               │
│ dashboard │ testing_data/dashboards/Other/flows-by-science-discipline.json    │
│ dashboard │ testing_data/dashboards/General/individual-flows.json             │
│ dashboard │ testing_data/dashboards/General/individual-flows-per-country.json │
│ dashboard │ testing_data/dashboards/Ignored/latency-patterns.json             │
│ dashboard │ testing_data/dashboards/General/loss-patterns.json                │
│ dashboard │ testing_data/dashboards/General/other-flow-stats.json             │
│ dashboard │ testing_data/dashboards/General/science-discipline-patterns.json  │
│ dashboard │ testing_data/dashboards/General/top-talkers-over-time.json        │
└───────────┴───────────────────────────────────────────────────────────────────┘
```

The dashboards will have a reference to the library panel linked by UID.

Here's the json from the dashboard JSON:

```json
      "libraryPanel": {
        "description": "",
        "meta": {
          "connectedDashboards": 3,
          "created": "2022-05-17T19:35:06Z",
          "createdBy": {
            "avatarUrl": "/avatar/579fc54abdc9ab34fb4865322f2870a1",
            "id": 13,
            "name": "mike.johnson@grafana.com"
          },
          "folderName": "mj",
          "folderUid": "R0bMCcW7z",
          "updated": "2022-05-17T19:37:14Z",
          "updatedBy": {
            "avatarUrl": "/avatar/579fc54abdc9ab34fb4865322f2870a1",
            "id": 13,
            "name": "mike.johnson@grafana.com"
          }
        },
        "name": "Extreme Dashboard Makeover - Grill",
        "type": "graph",
        "uid": "y1C0A5unz",
        "version": 2
      },
```

Please note, this is the Grill panel.

```json
{
         "name": "Extreme Dashboard Makeover - Grill",
        "orgId": 1,
        "type": "graph",
        "uid": "y1C0A5unz",
        "version": 1
}
```
## Deleting Elements
If we try to delete all the Library elements, that won't be allowed.

```sh
./bin/gdg lib clear
ERRO[0000] Failed to delete library panel titled: Dashboard Makeover - Extra Cleaning Duty Assignment Today  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Dashboard Makeover - Lighting Status  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Dashboard Makeover - Side Dish Prep Times, past 7 days  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Dashboard Makeover - Time since we purchased these spices  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Extreme Dashboard Makeover - Grill  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Extreme Dashboard Makeover - Mac Oven  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Extreme Dashboard Makeover - Refrigerator Temperature (F)  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Extreme Dashboard Makeover - Room Temperature (F)  ErrorMessage="the library element has connections"
ERRO[0000] Failed to delete library panel titled: Extreme Dashboard Makeover - Salmon Cooking Times, past 7 days  ErrorMessage="the library element has connections"
INFO[0000] No library were found.  0 librarys removed
```

### Deleting related dashboard


(Future version will allow you to inspect which dashboard has a link to which dashboards)

```sh
./bin/gdg dash clear -d dashboard-makeover-challenge                                                                                  (gke_esnet-sd-dev_us-central1-c_dev-staging-kafka-3/default)
INFO[0000] 1 dashboards were deleted
┌───────────┬──────────────────────────────┐
│ TYPE      │ FILENAME                     │
├───────────┼──────────────────────────────┤
│ dashboard │ Dashboard Makeover Challenge │
└───────────┴──────────────────────────────┘
```

Please note the -d, we're explicitly only deleting one dashboard.  We can verify the list.

```sh
./bin/gdg dash list
┌────┬──────────────────────────────┬──────────────────────────────┬─────────┬───────────┬────────────────────────────────────────────────────────────────┐
│ ID │ TITLE                        │ SLUG                         │ FOLDER  │ UID       │ URL                                                            │
├────┼──────────────────────────────┼──────────────────────────────┼─────────┼───────────┼────────────────────────────────────────────────────────────────┤
│ 80 │ Bandwidth Dashboard          │ bandwidth-dashboard          │ General │ 000000003 │ http://localhost:3000/d/000000003/bandwidth-dashboard          │
│ 81 │ Bandwidth Patterns           │ bandwidth-patterns           │ General │ 000000004 │ http://localhost:3000/d/000000004/bandwidth-patterns           │
│ 90 │ Flow Analysis                │ flow-analysis                │ Other   │ VuuXrnPWz │ http://localhost:3000/d/VuuXrnPWz/flow-analysis                │
│ 91 │ Flow Data for Circuits       │ flow-data-for-circuits       │ Other   │ xk26IFhmk │ http://localhost:3000/d/xk26IFhmk/flow-data-for-circuits       │
│ 92 │ Flow Data for Projects       │ flow-data-for-projects       │ Other   │ ie7TeomGz │ http://localhost:3000/d/ie7TeomGz/flow-data-for-projects       │
│ 93 │ Flow Data per Country        │ flow-data-per-country        │ Other   │ fgrOzz_mk │ http://localhost:3000/d/fgrOzz_mk/flow-data-per-country        │
│ 94 │ Flow Data per Organization   │ flow-data-per-organization   │ Other   │ QfzDJKhik │ http://localhost:3000/d/QfzDJKhik/flow-data-per-organization   │
│ 95 │ Flow Information             │ flow-information             │ Other   │ nzuMyBcGk │ http://localhost:3000/d/nzuMyBcGk/flow-information             │
│ 96 │ Flows by Science Discipline  │ flows-by-science-discipline  │ Other   │ WNn1qyaiz │ http://localhost:3000/d/WNn1qyaiz/flows-by-science-discipline  │
│ 83 │ Individual Flows             │ individual-flows             │ General │ -l3_u8nWk │ http://localhost:3000/d/-l3_u8nWk/individual-flows             │
│ 82 │ Individual Flows per Country │ individual-flows-per-country │ General │ 80IVUboZk │ http://localhost:3000/d/80IVUboZk/individual-flows-per-country │
│ 88 │ Latency Patterns             │ latency-patterns             │ Ignored │ 000000005 │ http://localhost:3000/d/000000005/latency-patterns             │
│ 84 │ Loss Patterns                │ loss-patterns                │ General │ 000000006 │ http://localhost:3000/d/000000006/loss-patterns                │
│ 85 │ Other Flow Stats             │ other-flow-stats             │ General │ CJC1FFhmz │ http://localhost:3000/d/CJC1FFhmz/other-flow-stats             │
│ 86 │ Science Discipline Patterns  │ science-discipline-patterns  │ General │ ufIS9W7Zk │ http://localhost:3000/d/ufIS9W7Zk/science-discipline-patterns  │
│ 87 │ Top Talkers Over Time        │ top-talkers-over-time        │ General │ b35BWxAZz │ http://localhost:3000/d/b35BWxAZz/top-talkers-over-time        │
└────┴──────────────────────────────┴──────────────────────────────┴─────────┴───────────┴────────────────────────────────────────────────────────────────┘
```

### Removing related components


```sh
./bin/gdg lib clear                                                                                                                   (gke_esnet-sd-dev_us-central1-c_dev-staging-kafka-3/default)
INFO[0000] 9 library were deleted
┌─────────┬────────────────────────────────────────────────────────────────┐
│ TYPE    │ FILENAME                                                       │
├─────────┼────────────────────────────────────────────────────────────────┤
│ library │ Dashboard Makeover - Extra Cleaning Duty Assignment Today      │
│ library │ Dashboard Makeover - Lighting Status                           │
│ library │ Dashboard Makeover - Side Dish Prep Times, past 7 days         │
│ library │ Dashboard Makeover - Time since we purchased these spices      │
│ library │ Extreme Dashboard Makeover - Grill                             │
│ library │ Extreme Dashboard Makeover - Mac Oven                          │
│ library │ Extreme Dashboard Makeover - Refrigerator Temperature (F)      │
│ library │ Extreme Dashboard Makeover - Room Temperature (F)              │
│ library │ Extreme Dashboard Makeover - Salmon Cooking Times, past 7 days │
└─────────┴────────────────────────────────────────────────────────────────┘
```
