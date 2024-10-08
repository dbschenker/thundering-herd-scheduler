# thundering-herd-scheduler

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.1.1](https://img.shields.io/badge/AppVersion-v0.1.1-informational?style=flat-square)

A Helm chart for Thundering herd scheduler

## Values

| Key | Type | Default                                                                                                                   | Description |
|-----|------|---------------------------------------------------------------------------------------------------------------------------|-------------|
| affinity | object | `{}`                                                                                                                      | Afinity for pods |
| fullnameOverride | string | `""`                                                                                                                      | Full name override |
| image | object | `{"pullPolicy":"IfNotPresent","repository":"ghcr.io/dbschenker/thundering-herd-scheduler","tag":"v1.29-0"}`               | Thundering-herd-scheduler container image settings |
| image.pullPolicy | string | `"IfNotPresent"`                                                                                                          | Image pull policy |
| image.repository | string | `"ghcr.io/dbschenker/thundering-herd-scheduler"`                                                                          | Registry address |
| image.tag | string | `"v1.29-0"`                                                                                                               | Image tag. Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]`                                                                                                                      | Map with names of image pull secrets |
| nameOverride | string | `""`                                                                                                                      | Name override |
| nodeSelector | object | `{}`                                                                                                                      | Node selector |
| podAnnotations | object | `{"prometheus.io/port":"10251","prometheus.io/scheme":"http","prometheus.io/scrape":"true"}`                              | Pod annotations |
| podDisruptionBudget.enabled | bool | `true`                                                                                                                    | Controls if PodDisruptionBadget object is created |
| podDisruptionBudget.minAvailable | int | `1`                                                                                                                       | Pod disruption budget - minAvailable. Enforces that at least one pod is available. |
| podSecurityContext | object | `{}`                                                                                                                      | Pod securoty context |
| replicaCount | int | `3`                                                                                                                       | Thundering-herd-scheduler replica count. By default it is set to 3. |
| resources | object | `{"limits":{"cpu":"250m","memory":"768Mi"},"requests":{"cpu":"100m","memory":"300Mi"}}`                                   | Resource limit and request settings |
| scheduler.burst | int | `60`                                                                                                                      | burst rate limiter setting |
| scheduler.logLevel | int | `1`                                                                                                                       | Thundering-herd-scheduler logging level |
| scheduler.pluginConfig.maxRetries | int | `5`                                                                                                                       | How many times a pod can run through the process before it anyway get's scheduled |
| scheduler.pluginConfig.parallelStartingPodsPerNode | int | `3`                                                                                                                       | How many pods should get scheduled in parallel before pods are moved into waiting state |
| scheduler.pluginConfig.timeoutSeconds | int | `5`                                                                                                                       | Based on how many times the pod was attempted to be scheduled using the scheduler, a wait is implemented with the following rule timeoutSeconds^2 * retries |
| scheduler.qps | int | `30`                                                                                                                      | qps rate limiter setting |
| securityContext | object | `{"capabilities":{"drop":["ALL"]},"privileged":false,"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1000}` | Security context settings |
| serviceAccount.annotations | object | `{}`                                                                                                                      | Annotations to add to the service account |
| serviceAccount.create | bool | `true`                                                                                                                    | Specifies whether a service account should be created |
| serviceAccount.name | string | `""`                                                                                                                      | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]`                                                                                                                      | Tolerations |
| topologySpreadConstraints | object | `{}`                                                                                                                      | Pod's topology spread constraint settings. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.8.1](https://github.com/norwoodj/helm-docs/releases/v1.8.1)
