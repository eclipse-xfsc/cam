# Remote Integrity Collection Module

A collection module to gather information about the integrity of the software stack used by a remote service.
It requests an Attestation Report from the remote service which includes

* Measurements from a trust anchor used by the remote service as well as
* Metadata required to describe the utilized software and
* All necessary certificates to validate the provided information.

The collection module requires an interface on the remote service which provides this report on request.
The result of the validation is passed on to the Evaluation Manager.

The module builds on the following existing tool for integrity validation: https://github.com/Fraunhofer-AISEC/cmc

## Necessary Information for Operation

- Information about the interface on the remote service responsible for providing the Attestation Report
- A (list of) trusted root certificate(s) utilized to validate the correctness of the Attestation Report
