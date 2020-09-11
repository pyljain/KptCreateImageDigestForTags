# kpt function to convert image tag to digests

Best practice is to use image digests in production and not image tags, as this
ensures that you are deploying the correct image in your workflows.

This kpt function scans your Deployment and Pod manifests to check if you are
using image tags. For public images, it fetches the image digest to make sure
that images are used with their digests and not tags.

```bash
go build && kpt fn run sample/ --enable-exec --exec-path ./kpt-sha-image
```
