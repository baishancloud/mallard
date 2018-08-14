a monitor system modified from open-falcon

in highly development, should not use in production

[![Build Status](https://travis-ci.org/baishancloud/mallard.svg?branch=master)](https://travis-ci.org/baishancloud/mallard)

[![codecov](https://codecov.io/gh/baishancloud/mallard/branch/master/graph/badge.svg)](https://codecov.io/gh/baishancloud/mallard)

### Introduction

The **mallard2** is distributed servers monitor system on baishancloud.com to monitor and discover breaks and problems for servers and services.

The first version is named *mallard* and based on open-falcon. After grant modification from open-falcon, it renames to the second version **mallard2**.

The **mallard2** includes some components to handle different functionals:

- mallard2-agent        : a metrics collector, collect server and service metrics data, run in each machines, handle raw metrics to simple events with judger strategies
- mallard2-transfer     : a gateway of data from agents, receive metrics and events data, dispatch metrics data to mallard2-store to save, dispatch events data to mallard2-eventor
- mallard2-store        : a gateway to saving data to influxdb
- mallard2-judge        : a multiple metrics judge unit
- malalrd2-eventor      : an aggregator for events, filter all events to find alarm events then send to mallard2-alarm
- mallard2-alarm        : an alarming to send alarm message to scripts or remote api
- mallard2-center       : a center to generate all configurations for transfer, eventor and alarm

### Installation


### Thanks

thank [open-falcon](https://github.com/open-falcon/falcon-plus) a lot