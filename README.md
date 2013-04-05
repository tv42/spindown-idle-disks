spindown-idle-disks -- Spin down idle SATA disks
================================================

Because `hdparm -S 120 /dev/sde` just won't work.

Use like this:

    sudo spindown-idle-disks -idle=10m /dev/sdd /dev/sde /dev/sdf

Leave running. Best to run it from Upstart or some such daemon
supervisor.
