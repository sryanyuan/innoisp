# innoisp

Innodb inspector, show the innodb table space file segments and pages information.

Just for learning innodb storage format.

## command

### overview

Overview the innodb table space file:

```innoisp overview -f db.ibd```

This command will show the table space file overview, following content will be output.

    ==========PAGE 0==========
    page num 0, offset 0x00000000, page type <File space header>
                    File header:
    Type <8> Checksum <0x4A13803A> Offset <0> Prev <0x00000000> Next <0x00000000> Log sequence number <1204662235189> Space ID <4202>
                    Page header:
    Heap top <0x0000> N heap <0x0000> Free <0x0000> Garbage <0x0000> Last insert <0x0000> Direction <0x0000> N direction <0x0000> N recs <0x0000> Index id <0x0000000000000000> Leaf inode <0x00000000:0x0000> Non-leaf inode <0x00000000:0x0000>
                    File trailer:
    Check sum<0x593CFE62> LSN<2071392309>

    ==========PAGE 1==========
    page num 1, offset 0x00004000, page type <Insert Buffer bit map>
                    File header:
    Type <5> Checksum <0x510BC8E0> Offset <1> Prev <0x00000000> Next <0x00000000> Log sequence number <1241148196> Space ID <4202>
                    Page header:
    Heap top <0x0000> N heap <0x0000> Free <0x0000> Garbage <0x0000> Last insert <0x0000> Direction <0x0000> N direction <0x0000> N recs <0x0000> Index id <0x0000000000000000> Leaf inode <0x00000000:0x0000> Non-leaf inode <0x00000000:0x0000>
                    File trailer:
    Check sum<0x203DCEE2> LSN<1241148196>

    ==========PAGE 2==========
    page num 2, offset 0x00008000, page type <File segment inode>
                    File header:
    Type <3> Checksum <0x259DFC10> Offset <2> Prev <0x00000000> Next <0x00000000> Log sequence number <1204662235189> Space ID <4202>
                    Page header:
    Heap top <0x0000> N heap <0x0000> Free <0x0000> Garbage <0x0000> Last insert <0x0000> Direction <0x0000> N direction <0x0000> N recs <0x0000> Index id <0x0000000000000000> Leaf inode <0x00000000:0x0000> Non-leaf inode <0x00000000:0x0000>
                    File trailer:
    Check sum<0x0272F708> LSN<2071392309>

    ==========PAGE 3==========
    page num 3, offset 0x0000C000, page type <Index> level <2>
                    File header:
    Type <17855> Checksum <0x587D68EB> Offset <3> Prev <0xFFFFFFFF> Next <0xFFFFFFFF> Log sequence number <1204662235189> Space ID <4202>
                    Page header:
    Heap top <0x0092> N heap <0x8004> Free <0x0000> Garbage <0x0000> Last insert <0x008A> Direction <0x0002> N direction <0x0001> N recs <0x0002> Index id <0x000000000000110D> Leaf inode <0x00000002:0x00F2> Non-leaf inode <0x00000002:0x0032>
                    File trailer:
    Check sum<0x598B926C> LSN<2071392309>
                    Page directory slots (2 total):
    [0x0070 0x0063]

### dslots

This command will show the index page's directory slot

```innoisp dslots -f db.ibd```

                            ==========PAGE 3 OFFSET 0xC000 LEVEL 2==========
    slot    offset      type        owned   key
    0       0x0063      infimum     1       N/A
    slot reference: [infimum own 1 0x0063]
    1       0x0070      supremum    3       N/A
    slot reference: [0x007D K0]->[0x008A K0]->[supremum own 3 0x0070]

    ........................................

### space

This command will show the file space header page and extend descriptor page information

```innoisp space -f db.ibd```

                            ==========PAGE 0 OFFSET 0x0000==========
    space id  page allo  page init  flags   page used(fg)  free_frag list                                     free list                                          full_frag list                                     next segment id  full inodes                                        free inodes
    4202      2304       1664       0x0000  38             len<1> 0x00000000:0x0096 0x00000000:0x0096         len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         3                len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         len<1> 0x00000002:0x0000 0x00000002:0x0000

    extend       page range          file segment id     state           list
    0(0x0096)    0-63                0x0000000000000000  0x00000002      0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000
    1(0x00BE)    64-127              0x0000000000000002  0x00000004      0xFFFFFFFF:0x0000 0x00000000:0x00E6
    2(0x00E6)    128-191             0x0000000000000002  0x00000004      0x00000000:0x00BE 0x00000000:0x010E
    3(0x010E)    192-255             0x0000000000000002  0x00000004      0x00000000:0x00E6 0x00000000:0x0136
    4(0x0136)    256-319             0x0000000000000002  0x00000004      0x00000000:0x010E 0x00000000:0x015E
    5(0x015E)    320-383             0x0000000000000002  0x00000004      0x00000000:0x0136 0x00000000:0x0186
    6(0x0186)    384-447             0x0000000000000002  0x00000004      0x00000000:0x015E 0x00000000:0x01AE
    7(0x01AE)    448-511             0x0000000000000002  0x00000004      0x00000000:0x0186 0x00000000:0x01D6
    8(0x01D6)    512-575             0x0000000000000002  0x00000004      0x00000000:0x01AE 0x00000000:0x01FE
    9(0x01FE)    576-639             0x0000000000000002  0x00000004      0x00000000:0x01D6 0x00000000:0x0226
    10(0x0226)   640-703             0x0000000000000002  0x00000004      0x00000000:0x01FE 0x00000000:0x024E
    11(0x024E)   704-767             0x0000000000000002  0x00000004      0x00000000:0x0226 0x00000000:0x0276
    12(0x0276)   768-831             0x0000000000000002  0x00000004      0x00000000:0x024E 0x00000000:0x029E
    13(0x029E)   832-895             0x0000000000000002  0x00000004      0x00000000:0x0276 0x00000000:0x02C6
    14(0x02C6)   896-959             0x0000000000000002  0x00000004      0x00000000:0x029E 0x00000000:0x02EE
    15(0x02EE)   960-1023            0x0000000000000002  0x00000004      0x00000000:0x02C6 0x00000000:0x0316
    16(0x0316)   1024-1087           0x0000000000000002  0x00000004      0x00000000:0x02EE 0x00000000:0x033E
    17(0x033E)   1088-1151           0x0000000000000002  0x00000004      0x00000000:0x0316 0x00000000:0x0366
    18(0x0366)   1152-1215           0x0000000000000002  0x00000004      0x00000000:0x033E 0x00000000:0x038E
    19(0x038E)   1216-1279           0x0000000000000002  0x00000004      0x00000000:0x0366 0x00000000:0x03B6
    20(0x03B6)   1280-1343           0x0000000000000002  0x00000004      0x00000000:0x038E 0x00000000:0x03DE
    21(0x03DE)   1344-1407           0x0000000000000002  0x00000004      0x00000000:0x03B6 0x00000000:0x0406
    22(0x0406)   1408-1471           0x0000000000000002  0x00000004      0x00000000:0x03DE 0x00000000:0x042E
    23(0x042E)   1472-1535           0x0000000000000002  0x00000004      0x00000000:0x0406 0x00000000:0x0456
    24(0x0456)   1536-1599           0x0000000000000002  0x00000004      0x00000000:0x042E 0xFFFFFFFF:0x0000
    25(0x047E)   1600-1663           0x0000000000000002  0x00000004      0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000


### inode

This file will show the file segment inode page information

```innoisp inode -f db.ibd```

                            ==========PAGE 2 OFFSET 0x8000==========
    page list
    0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000

    file segment id     used(nf)  free list                                          not_full list                                      full list                                          fragment array
    0x00000032:1        0         len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         3 36 37 (page allocate)
    0x000000F2:2        36        len<0> 0xFFFFFFFF:0x0000 0xFFFFFFFF:0x0000         len<1> 0x00000000:0x047E 0x00000000:0x047E         len<24> 0x00000000:0x00BE 0x00000000:0x0456        4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 (extend allocate)

## TODO list

### search

Search the ibd file through index non-leaf page