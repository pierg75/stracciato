# Stracciato

It's a little tool that analyses a strace output file.
It's something I often need and I wanted to write more Go, so here it is.

The strace file needs to be in a specific format, ideally from the command:
```
% strace -fttTvxyo <output file> <command>
```

To use `stracciato`, simply download this repo, either compile it or run it directly.
For example:
```
% go build .
```

```
% ./stracciato tests/strace
Syscall              Count      Average    Min        Max        Total
read                 4          0.000020   0.000019   0.000021   0.000080
statfs               2          0.000025   0.000021   0.000030   0.000051
getrandom            1          0.000025   0.000025   0.000025   0.000025
execve               1          0.000620   0.000620   0.000620   0.000620
openat               7          0.000032   0.000026   0.000040   0.000227
fstat                8          0.000018   0.000017   0.000020   0.000147
close                9          0.000027   0.000018   0.000078   0.000243
munmap               1          0.000046   0.000046   0.000046   0.000046
prctl                6          0.000016   0.000015   0.000017   0.000097
exit_group           0          0.000000   0.000000   0.000000   0.000000
brk                  3          0.000026   0.000017   0.000031   0.000078
access               2          0.000029   0.000022   0.000037   0.000059
mmap                 22         0.000029   0.000024   0.000042   0.000636
set_tid_address      1          0.000015   0.000015   0.000015   0.000015
rseq                 1          0.000017   0.000017   0.000017   0.000017
ioctl                2          0.000022   0.000019   0.000024   0.000043
arch_prctl           1          0.000017   0.000017   0.000017   0.000017
set_robust_list      1          0.000017   0.000017   0.000017   0.000017
mprotect             6          0.000026   0.000023   0.000032   0.000155
write                1          0.000030   0.000030   0.000030   0.000030

Number of threads: 1

There were some unknown calls, use '--verbose/-v' to see them.
```
