# import-lastpass

A golang version of
the
[Ruby](https://git.zx2c4.com/password-store/tree/contrib/importers/lastpass2pass.rb) version.

The Ruby did a simple split on the comma separators in each
line. However, Lasspass allows for safe notes (in German: Sichere
Notizen), which can of course be multi line and which have `http://sn`
as in the `url` field. These multi line `extra` fields were not
correctly handled by the Ruby version.

The `encoding/csv` from the standard Go library handles these
Lastpass export csv files just fine, hence this program.

You will have to have the two Go packages

   	github.com/mkideal/cli
	github.com/pkg/errors

installed.
