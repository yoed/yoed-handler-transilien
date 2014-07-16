Yo'ed transilien
======================

Yo'ed handler sending back a Yo if a train is on time on the Paris Transilien network

#Installation
You need the [go](http://golang.org) package on your machine to get the source

`go get github.com/yoed/yoed-client-transilien`

#Configuration
Create a `config.json` file aside the executable program.
For more informations about the basic configuration, [see](https://github.com/yoed/yoed-client-interface#yoed-client-interface)

#FromStation (string)
The departure station code.

#ToStation (string)
The arrival station code

#Hour (configTime)
The hour when your train is

#Delta (configDelta)
The allowed delta (in case the train is delayed) with the original hour

