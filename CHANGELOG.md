## [Omni Infra Provider KubeVirt 0.3.0](https://github.com/siderolabs/omni-infra-provider-kubevirt/releases/tag/v0.3.0) (2026-09-16)

Welcome to the v0.3.0 release of Omni Infra Provider KubeVirt!



Please try out the release binaries and report any issues at
https://github.com/siderolabs/omni-infra-provider-kubevirt/issues.

### Authenticated Image Factory Support

Previously, the provider built the download URL of the boot image itself, using the public image factory.
An Omni configured with an image factory that authenticates its downloads never got a usable boot image this way.

The provider now asks Omni for the URL and for the headers which go with it. The headers are passed on to the importer which downloads the image.


### Contributors

* Utku Ozdemir
* Oguz Kilcan

### Changes
<details><summary>3 commits</summary>
<p>

* [`a037ae6`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/a037ae6543f8b1782816f82bdad1fb322fcd88da) test: run the integration tests against the enterprise image factory
* [`125ad00`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/125ad00915915bb23a7f1e1799c572d753d466bb) chore: rekres
* [`6f582cc`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/6f582cc51d621e3fa83bffa2a31b422df5005643) feat: support authenticated image factory downloads
</p>
</details>

### Dependency Changes

* **github.com/cosi-project/runtime**              v1.16.2 -> v1.16.3
* **github.com/go-logr/logr**                      v1.4.3 -> v1.4.4
* **github.com/planetscale/vtprotobuf**            ba97887b0a25 -> 8ae5a48058df
* **github.com/siderolabs/omni/client**            582730ce940c -> e460ae71eaae
* **google.golang.org/protobuf**                   f2248ac996af -> v1.36.12
* **k8s.io/api**                                   v0.36.3 -> v0.37.0
* **k8s.io/apimachinery**                          v0.36.3 -> v0.37.0
* **k8s.io/client-go**                             v0.36.3 -> v0.37.0
* **kubevirt.io/api**                              v1.8.2 -> v1.9.0
* **kubevirt.io/containerized-data-importer-api**  v1.65.0 -> v1.66.0

Previous release can be found at [v0.2.0](https://github.com/siderolabs/omni-infra-provider-kubevirt/releases/tag/v0.2.0)

## [Omni Infra Provider KubeVirt 0.2.0](https://github.com/siderolabs/omni-infra-provider-kubevirt/releases/tag/v0.2.0) (2026-07-23)

Welcome to the v0.2.0 release of Omni Infra Provider KubeVirt!



Please try out the release binaries and report any issues at
https://github.com/siderolabs/omni-infra-provider-kubevirt/issues.

### Add Versioning Support

The infra provider will now report its version to Omni.


### Contributors

* Edward Sammut Alessi

### Changes
<details><summary>1 commit</summary>
<p>

* [`7081602`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/70816020eb82f8036183999968e2c0390270c2dc) feat: rekres and add versioning
</p>
</details>

### Dependency Changes

* **github.com/cosi-project/runtime**    v1.16.1 -> v1.16.2
* **github.com/siderolabs/omni/client**  v1.8.0 -> 582730ce940c
* **k8s.io/api**                         v0.36.1 -> v0.36.3
* **k8s.io/apimachinery**                v0.36.1 -> v0.36.3
* **k8s.io/client-go**                   v0.36.1 -> v0.36.3

Previous release can be found at [v0.1.0](https://github.com/siderolabs/omni-infra-provider-kubevirt/releases/tag/v0.1.0)

## [Omni KubeVirt Infra Provider 0.1.0](https://github.com/siderolabs/omni-infra-provider-kubevirt/releases/tag/v0.1.0) (2026-05-28)

Welcome to the v0.1.0 release of Omni KubeVirt Infra Provider!



Please try out the release binaries and report any issues at
https://github.com/siderolabs/omni-infra-provider-kubevirt/issues.

### Contributors

* Artem Chernyshev
* Farenjihn
* Andrey Smirnov
* Artem Chernyshev
* Ganawa Juanah
* Maxime
* Niklas Voss
* Ryan King
* Spencer Smith

### Changes
<details><summary>25 commits</summary>
<p>

* [`67feb41`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/67feb41b381db470cf8a7db7596f35aa56a88916) chore: bump dependencies
* [`8e4ba0f`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/8e4ba0f0305271087b22df5533a15316fb5a3a08) chore: bump deps
* [`c83f4d0`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/c83f4d0aae4677c71a4b504a58153415c491c5dc) feat: add vm_labels support to provider data
* [`25af68f`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/25af68fa3adafb6b667f0aab66f322dc041ddb16) chore: update crypto import
* [`d7de6b8`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/d7de6b8ac0f3bee1209c1a1ebce7249e83197d36) refactor: improve readability of in-cluster fallback
* [`5c26277`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/5c26277b664ccf01e2948e764c031c79a18bdfde) feat: use in-cluster client as fallback
* [`142adcf`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/142adcfcba798c98366e46e4f4242061c8725749) feat: allow using in-cluster config for kubernetes client
* [`9db964d`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/9db964d90da960efeac1be33f2217cee16fdbc63) feat: add tolerations to data schema
* [`f196ca4`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/f196ca49fe19240834e7c4da6ac8a193209c091d) feat: add storageclassname to data schema
* [`e71d788`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/e71d78868d262de010d1190c2d7b845f5cb4e640) fix: add missing flags to the Omni executable in the integration tests
* [`b023994`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/b023994359e6934a7e36a05e0dcaae92feb67b41) chore: bump deps and rekres
* [`b379dab`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/b379dab3c230102edb979f4e4ffa141866e4e51e) refactor: update and rekres with new QRuntime
* [`4d44251`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/4d44251e64fea2824bedb37af2968f97d2a4a607) fix: change DataVolume size to fix too small to contain image errors
* [`1bf75bb`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/1bf75bb144e3751dd087ed0a7c80fa2f14c5b6d0) chore: update Omni client and other Go deps
* [`a83c5f4`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/a83c5f451638bc98f2d806c811c5db40712d2697) feat: encode the request IDs into Omni join tokens
* [`20e1641`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/20e1641a2c024355d221b0cf1c81d510582a4ca4) chore: bump deps and rekres
* [`5c788c9`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/5c788c9850e17ff2077f48dfacac89723de18b35) chore: bump deps
* [`1897fe5`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/1897fe5d9dc144aebdac44e63d004b7be4c787dc) docs: fix commands in readme.md
* [`39f7c7c`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/39f7c7cb1b22abc0260e758f6fa3e753c4d72a9a) test: use infra provider service account in the tests
* [`fc6f563`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/fc6f5633022af761f888778e4f2b54a7d7c5573c) fix: allow using custom provider ID
* [`6e22e7f`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/6e22e7fd72a5f68bebe445c8e8c91b513f51cc53) feat: configure network binding through provider data instead
* [`deb6fcc`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/deb6fcc46a3f63a87adf652c1362ac482cd010d8) test: add integration tests
* [`8c4fc0b`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/8c4fc0ba3971d89ab1ccc5f40e96067624c16691) feat: implement KubeVirt provider
* [`95de1fe`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/95de1fe9b9c910026071987fa97362a3ebf65f80) chore: add GHA
* [`94474bc`](https://github.com/siderolabs/omni-infra-provider-kubevirt/commit/94474bc6de6ad91419cee6fba863db6a8f142ba6) Initial commit
</p>
</details>

### Dependency Changes

This release has no dependency changes

