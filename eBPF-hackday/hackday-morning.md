 1675  # ----- BEGIN eBPF HACKDAY
 1676  git clone git@github.com:iovisor/bpftrace.git
 1677  mkdir -p bpftrace/build
 1678  cd bpftrace/build
 1679  cmake
 1680  # zsh: command not found: cmake
 1681  #< zsh: command not found: cmake
 1682  brew search cmake
 1683  brew install cmake
 1684  cmake -DCMAKE_BUILD_TYPE=Release ../
 1685  #< CMake Error
 1686  #< Please install the bcc library package, which is required.
 1687  brew search bcc
 1688  #> this has an aweful lot of linux stuff...
 1689  git clone https://github.com/iovisor/bcc.git\nmkdir bcc/build; cd bcc/build\ncmake ..\nmake\nsudo make install
 1690  #< CMake Error
 1691  #< Could not find a package configuration file provided by "LLVM" with any of\n  the following names:\n\n    LLVMConfig.cmake\n    llvm-config.cmake
 1692  #< Could not find a package configuration file provided by "LLVM" with any of\n#<  the following names:\n\n#<    LLVMConfig.cmake\n#<    llvm-config.cmake
 1693  j bpft
 1694  ll
 1695  cd bpftrace
 1696  ll
 1697  cd -
 1698  ll
 1699  rm -rf bpftrace
 1700  git clone https://github.com/iovisor/bpftrace\nmkdir -p bpftrace/build\ncd bpftrace/build\ncmake -DCMAKE_BUILD_TYPE=Release ../\nmake
 1701  cd -
 1702  git clone https://github.com/iovisor/bcc.git\nmkdir bcc/build; cd bcc/build\ncmake ..\nmake\nsudo make install
 1703  cd -
 1704  #> this feels like I'm going down a terrible rabbithole
 1705  brew search llvm
 1706  brew install llvm
 1707  #< llvm is keg-only, which means it was not symlinked into /usr/local,\n#< because macOS already provides this software and installing another version in\n#< parallel can cause all kinds of trouble.
 1708  which llvm
 1709  #< llvm not found
 1710  echo 'export PATH="/usr/local/opt/llvm/bin:$PATH"' >> ~/.zshrc
 1711  source ~/.zshrc
 1712  llvm
 1713  ll /usr/local/opt/llvm
 1714  ll /usr/local/opt/llvm/bin
 1715  which llvm-cat 
 1716  #> oh, so I do have llvm stuff
 1717  #> here's where it would have been nice to have set some kind of marker at the beginning of this rabbit hole and been able to retrieve it with one simple command
 1718  ll
 1719  rm -rf bcc
 1720  git clone https://github.com/iovisor/bcc.git\nmkdir bcc/build; cd bcc/build\ncmake ..\nmake\nsudo make install
 1721  #< CMake Error
 1722  #< Could NOT find LibElf
 1723  #> deer lord
 1724  history | grep BEGIN eBPF
 1725  history | grep 'BEGIN eBPF'
 1726  history | grep 'BEGIN eBPF HACKDAY'
 1727  history | grep 'BEGIN eBPF HACKDAY' | cut 1
 1728  history | grep 'BEGIN eBPF HACKDAY' | awk '{print $1}'
 1729  history 25
 1730  history 15
 1731  history | tail 20
 1732  history | tail -n20
 1733  history | tail -n50
 1734  history | tail -n75
