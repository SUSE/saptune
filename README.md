
[![Build Status](https://travis-ci.org/SUSE/saptune.svg?branch=master)](https://travis-ci.org/SUSE/saptune)
[![Test Coverage](https://api.codeclimate.com/v1/badges/5375e2ca293dd0e8b322/test_coverage)](https://codeclimate.com/github/SUSE/saptune/test_coverage)
[![Maintainability](https://api.codeclimate.com/v1/badges/5375e2ca293dd0e8b322/maintainability)](https://codeclimate.com/github/SUSE/saptune/maintainability)


# saptune

# What is saptune?

If you have never heard about saptune, now is good time to explore it.

Saptune – part of SLES for SAP Applications – is a configuration tool to prepare a system to run SAP workloads by implementing the recommendations of various SAP notes. Just select the notes you need or choose one of the predefined groups – called solutions.

# Why saptune?

To get SAP applications work properly, a lot of system parameters need to set to specific values.\
These various parameter settings are mostly hidden in a bunch of SAP Notes.\
Often difficult to find for the customer and inconvenient to apply manually – mostly error-prone.\
Need for check, if a system fully conforms to the requirements of SAP (important in Cloud environments).

# So our goals with saptune are:

Provide a central framework for configuring the SLES for running various SAP workloads based on SAP recommendations

Enable partners and customers to extend saptune with their own configurations and tests


# Highlights:

Enhanced human-readable output so you can get a clear overview which SAP notes are available, which SAP notes are referred by a dedicated solution, which SAP note or solution is currently in use (aka applied).

A detailed output for verify and simulate which now tells everything you need to know about SAP notes and solutions.
One look is enough to see if your system fully conforms to the requirements of SAP.

No secrets - Implementing SAP note recommendations as fully as possible. Only where it is not safe to do so, saptune will just notify you without automatically implementing them like the modifications of the boot loader.

We haven’t a fully defined logging yet (a broad hint for the future), but with the current release saptune will record a lot more details.

Every parameter is now configurable, it is listed in the configuration file and can be overwritten or marked to be left alone by saptune.

More features available in extra configurations. Almost every configuration type saptune has to offer can be used in the extra configuration files too.

But what are these extra files for?\
They give our administrators and our partners a simple way to implement their own additional configurations. For example, this allows administrators to implement SAP Notes not yet shipped with saptune, or to centralize system configurations with saptune. No more need to spread configurations over multiple tools like sysctl.conf, limits.conf etc. If you run saptune, let it do this job! In short, you can have your “own SAP Note” to be applied by saptune.


# Migration:

Guide to help you migrate to the new saptune\
Migration? Isn’t a simple package update enough?

Not in every case.

We changed quite a lot and we don’t want to risk causing any incompatibilities or unexpected changes in your system behavior.\
If the update discovers applied saptune SAP notes or solutions, saptune will continue to run in version 1.\
The switch to version 2 has to be done deliberately.\
To help you, we will provide a step-by-step guide. Just plan your switch when you are ready, no rush!

We will support saptune version 1 until end of the lifetime of SLES 12 / SLES 15 SP1, which should give enough time to move. Although please bear in mind that since saptune version 1 will be deprecated, we will only do bug fixing. New features, new SAP notes or new parameters will only be done for version 2!


# Where to find documentation?

The saptune package will contain detailed man pages. You can find the pdf version of the man pages here in the repository in the doc directory.\
Also SAP note “1275776 – Linux: Preparing SLES for SAP environments” will get an update to reflect both versions.\
When the technical blog series about the details of saptune and how to do a migration from version 1 to version 2 will be available, the link collection will be updated.\
For now:\
<https://www.suse.com/c/a-new-saptune-is-knocking-on-your-door/>\
<https://www.suse.com/c/a-new-saptune-is-here/>\
<https://www.suse.com/c/saptune-a-deep-dive/>


# Feedback

Supporters, contributors, colleagues, customers and partners are welcome to approach us with ideas and suggestions. If you miss something or you think something should be done better, then don’t hesitate to contact us. You are welcome to give further feedback via email at SapAlliance@suse.com, create an issue in this repository, carrier pigeon etc. and tell us your needs.\
With each new version of saptune we implement many of them, but the journey will continue and you can expect further enhancements in the future.


Enjoy the new saptune!

---

[Hints for development](development.md)

