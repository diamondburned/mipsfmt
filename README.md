# mipsfmt

Automatically format your SPIM MIPS32 Assembly sources with a simple command.

mipsfmt is a fork of [diamondburned/nasmfmt](https://libdb.so/nasmfmt). The
rewrite supports SPIM MIPS32 Assembly instead of NASM.

Inspired by gofmt.

## Example

```
# Program to add two plus three 
.text
   .globl  main

main:
   ori  $8,$0,0x2   # put two's comp. two into register 8
  ori $9,$0,0x3 # put two's comp. three into register 9
        addu  $10,$8,$9  # add register 8 and 9, put result in 10

## End of file
```

becomes

```
## Program to add two plus three 
        .text  
        .globl main

main:
        ori  $8, $0, 0x2               # put two's comp. two into register 8
        ori  $9, $0, 0x3               # put two's comp. three into register 9
        addu $10, $8, $9               # add register 8 and 9, put result in 10

## End of file
```

## Installing

Requires Go 1.18+.

```go
go install libdb.so/mipsfmt@latest
```

## Vim + ALE integration

```vim
autocmd BufRead,BufNewFile *.s    set filetype=mips

function! FixMipsfmt(buffer) abort
    return {
    \   'command': 'mipsfmt -'
    \}
endfunction

execute ale#fix#registry#Add('mipsfmt', 'FixMipsfmt', ['mips'], 'mipsfmt')

let g:ale_fixers = {
	\ 'mips': [ "mipsfmt" ],
	\ }
```
