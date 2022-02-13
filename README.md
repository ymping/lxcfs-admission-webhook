# lxcfs-admission-webhook

<div id="top"></div>
<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->



<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![Apache License][license-shield]][license-url]


<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://linuxcontainers.org/lxcfs/introduction/">
    <img src="https://linuxcontainers.org/static/img/containers.small.png" alt="LXCFS Logo" width="80" height="80">
  </a>
  <strong>on</strong>
  <a href="https://kubernetes.io/">
    <img src="https://kubernetes.io/images/favicon.png" alt="K8s Logo" width="80" height="80">
  </a>

<h3 align="center">LXCFS Admission Webhook</h3>

  <p align="center">
    Correct the linux container's CGroup-view by <a href="https://linuxcontainers.org/lxcfs/introduction/">LXCFS</a> and <a href="https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/">kubernetes admission webhook</a>
    <br />
    <br />
    <a href="https://github.com/ymping/lxcfs-admission-webhook"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/ymping/lxcfs-admission-webhook/issues">Report Bug</a>
    ·
    <a href="https://github.com/ymping/lxcfs-admission-webhook/issues">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>



<!-- ABOUT THE PROJECT -->
## About The Project

[![Product Name Screen Shot][product-screenshot]](https://www.processon.com/view/link/6208d461f346fb3a0a38d972)

LXCFS is a simple userspace filesystem designed to work around some current limitations of the Linux kernel.

This project run LXCFS as kubernetes daemonset and expose a set of files which can be bind-mounted over kubernetes pod's /proc originals to provide CGroup-aware values.

Also provide a kubernetes admission webhook make the pod which need to correct their CGroup-view to auto bind-mount LXCFS files.

<p align="right">(<a href="#top">back to top</a>)</p>



### Built With

* [Golang](https://go.dev/)
* [LXCFS](https://linuxcontainers.org/lxcfs/introduction/)
* [Kubernetes Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
 
<p align="right">(<a href="#top">back to top</a>)</p>



<!-- GETTING STARTED -->
## Getting Started

This is an example of how you may give instructions on setting up your project locally.
To get a local copy up and running follow these simple example steps.

### Prerequisites

* Kubernetes version
  Kubernetes v1.16 or above with the `admissionregistration.k8s.io/v1` API enabled. Verify that by the following command:
  ```
  kubectl api-versions | grep admissionregistration.k8s.io/v1
  ```
  The result should be:
  ```
  admissionregistration.k8s.io/v1
  ```
  The API `admissionregistration.k8s.io/v1beta1` not tested, not recommended.

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/ymping/lxcfs-admission-webhook.git
   ```
3. Run install script
   
   Default install webhook and LXCFS service in kubernetes namespace `lxcfs`,
   use `install.sh --namespace your_ns` to deploy service in specify namespace.
   ```sh
   cd deploy
   ./install.sh
   ```
4. Go to [usage](#usage) section see how to usage
5. Uninstall

   Default uninstall webhook and LXCFS service in kubernetes namespace `lxcfs`,
   use `uninstall.sh --namespace your_ns` to uninstall service in specify namespace.
   ```sh
    cd deploy
   ./uninstall.sh
   ```

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- USAGE EXAMPLES -->
## Usage

1. Add lable `lxcfs-admission-webhook=enabled` to namespace which you want to correct the linux container's CGroup-view.
   ```sh
   kubectl label namespaces your_namespace lxcfs-admission-webhook=enabled
   ```
2. If you want disable this feature on some specify pod,
   add an annotation `mutating.lxcfs-admission-webhook.io/enable` to the pod,
   the webhook will skip patch this pod when create it.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

- [ ] Support helm installation

See the [open issues](https://github.com/ymping/lxcfs-admission-webhook/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTRIBUTING -->
## Contributing

Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- LICENSE -->
## License

Distributed under the Apache License 2.0. See `LICENSE` for more information.

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- CONTACT -->
## Contact

ymping - [ympiing@gmail.com](mailto:ympiing@gmail.com)

Project Link: [https://github.com/ymping/lxcfs-admission-webhook](https://github.com/ymping/lxcfs-admission-webhook)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* ['unknown revision v0.0.0' errors](https://github.com/kubernetes/kubernetes/issues/79384) during development,
  run `go-get-k8s-pkg.sh vXXX` to fix, `vXXX` is kubernetes version, example `go-get-k8s-pkg.sh v1.23.3`

* [how to create self signed certs](https://kubernetes.io/docs/tasks/administer-cluster/certificates/#openssl)

<p align="right">(<a href="#top">back to top</a>)</p>



<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ymping/lxcfs-admission-webhook.svg?style=for-the-badge
[contributors-url]: https://github.com/ymping/lxcfs-admission-webhook/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/ymping/lxcfs-admission-webhook.svg?style=for-the-badge
[forks-url]: https://github.com/ymping/lxcfs-admission-webhook/network/members
[stars-shield]: https://img.shields.io/github/stars/ymping/lxcfs-admission-webhook.svg?style=for-the-badge
[stars-url]: https://github.com/ymping/lxcfs-admission-webhook/stargazers
[issues-shield]: https://img.shields.io/github/issues/ymping/lxcfs-admission-webhook.svg?style=for-the-badge
[issues-url]: https://github.com/ymping/lxcfs-admission-webhook/issues
[license-shield]: https://img.shields.io/github/license/ymping/lxcfs-admission-webhook.svg?style=for-the-badge
[license-url]: https://github.com/ymping/lxcfs-admission-webhook/blob/master/LICENSE
[product-screenshot]: http://assets.processon.com/chart_image/6208c9970e3e7407d1cddc1d.png
