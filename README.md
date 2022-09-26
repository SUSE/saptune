
[![Build Status](https://github.com/SUSE/saptune/actions/workflows/saptune-ut.yml/badge.svg)](https://github.com/SUSE/saptune/actions/workflows/saptune-ut.yml/badge.svg)
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

What is a migration?

With saptune 2 we changed a lot of things and it was not possible to keep and use the configuration from the previous version 1. Manual steps are necessary.\
Therefore after an upgrade, saptune 2 run as version 1 and kept the previous configuration, if applied saptune SAP notes or solutions had been discovered.\
As a result the upgrade does not broke the system tuning and you had time to migrate the configuration afterwards.\
The man page saptune-migrate(7) contains a detailed guide what to do.

With saptune version 3 the support for version 1 is abandoned.\
If the migration from version 1 to version 2 or 3 was not done before the package update, saptune will not work and that's expected.\
But all needed files for the migration are still available, so following the procedure described in the man page saptune-migrate(7) will bring you back to a working saptune.

*Migration is only needed, if you still have a version 1 configuration! In all other cases you can simply upgrade the package, a migration is not necessary.*


# Moving away from tuned:

With saptune 3, we discontinued the use of 'tuned' and use a systemd service 'saptune.service' instead.

This switch makes it necessary to stop the tuning of the system during the package update for a short timeframe to avoid tuned error messages.

If you are used to start saptune using 'tuned' directly in the past, please move to 'saptune service start' instead.
If you use 'saptune daemon start', you will now get a 'deprecated' warning, but the command will still continue to work.


# Where to find documentation?

The saptune package will contain detailed man pages. You can find the pdf version of the man pages here in the repository in the doc directory.\
Also SAP note “1275776 – Linux: Preparing SLES for SAP environments” will get an update to reflect both versions.\
When the technical blog series about the details of saptune and how to do a migration from version 1 to version 2 will be available, the link collection will be updated.\
For now:\
<https://www.suse.com/c/help-saptune-says-my-system-is-degraded/>\
<https://www.suse.com/c/a-new-saptune-is-knocking-on-your-door/>\
<https://www.suse.com/c/a-new-saptune-is-here/>\
<https://www.suse.com/c/saptune-a-deep-dive/>\
<https://www.suse.com/c/saptune-3-is-on-the-horizon/>\
<https://www.suse.com/c/saptune-3-is-here/>


# Feedback

Supporters, contributors, colleagues, customers and partners are welcome to approach us with ideas and suggestions. If you miss something or you think something should be done better, then don’t hesitate to contact us. You are welcome to give further feedback via email at sapalliance@suse.com, create an issue in this repository, carrier pigeon etc. and tell us your needs.\
With each new version of saptune we implement many of them, but the journey will continue and you can expect further enhancements in the future.


Enjoy the new saptune!

---

[Hints for development](development.md)

