# YAMLMerge

Takes an arbitrary number of yaml files, extracts $path from each, and merges these into a template at a new $path

## Configuration

Create a new file `config.yaml` with the following:

- `template` String. The template to merge keys into
- `crds` List. The yaml files you wish to extract from

The structure for `crds` is an object as such:

- `mergeAt` The yaml path in the template you wish to merge to
- `mergeFrom` The yaml path in the remote file you need to extract
- `fileUrl` The URL to the remote file
- `version` The version to load

Example:

```yaml
template: ./xrd/definition.yaml
crds:
  - mergeAt: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.aws.properties.cluster
    mergeFrom: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.atProvider.properties
    fileUrl: "https://raw.githubusercontent.com/upbound/provider-aws/{{ .Version }}/package/crds/eks.aws.upbound.io_clusters.yaml"
    version: v0.43.0
  - mergeAt: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.aws.properties.vpc
    mergeFrom: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.atProvider.properties
    fileUrl: "https://raw.githubusercontent.com/upbound/provider-aws/{{ .Version }}/package/crds/ec2.aws.upbound.io_vpcs.yaml"
    version: v0.43.0
  - mergeAt: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.azure.properties.cluster
    mergeFrom: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.atProvider.properties
    fileUrl: "https://raw.githubusercontent.com/upbound/provider-azure/{{ .Version }}/package/crds/containerservice.azure.upbound.io_kubernetesclusters.yaml"
    version: v0.38.0
  - mergeAt: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.gcp.properties.cluster
    mergeFrom: .spec.versions[0].schema.openAPIV3Schema.properties.status.properties.atProvider.properties
    fileUrl: "https://raw.githubusercontent.com/upbound/provider-gcp//{{ .Version }}/package/crds/container.gcp.upbound.io_clusters.yaml"
    version: v0.38.0
```

Results in the structure:

```yaml
status:
  type: object
  properties:
    azure:
      type: object
      properties:
        cluster:
        type: object
        properties:
            ...
    aws:
      type: object
      properties:
        cluster:
        type: object
        properties:
            ...
        vpc:
        type: object
        properties:
            ...
    gcp:
      type: object
      properties:
        cluster:
        type: object
        properties:
            ...
```

The key `properties` is inserted if it does not exist, otherwise it is replaced.