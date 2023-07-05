# std

The standard library of Jule. <br>

+ The directories in this directory, accept as package.
+ The files in this directory, accept as builtin package files.

## Important

### Packages

Name of package directories is must conform to naming conventions of the language.

### Builtin Package Files

Builtin package files is actually direct imported defines. <br>
This files is can't imports in source code because before compilation, compiler imports these package automatically. <br>
So, not necessary. Actually, is not possible because ``use`` statement sees subpackages. <br>
For this reason, this defines, can use in everywhere and not necessary any import operation for use. <br>
