  $ spindown-idle-disks -idle=bork sdborkbork
  invalid value "bork" for flag -idle: time: invalid duration bork
  Usage of spindown-idle-disks:
    spindown-idle-disks [OPTS] DISK [DISK..]
    -idle=10m0s: how long disk needs to be idle
  [2]

  $ spindown-idle-disks does-not-exist
  spindown-idle-disks: no such device: stat does-not-exist: no such file or directory
  [1]
