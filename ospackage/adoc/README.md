# Building man pages

Run the `./build` command within this directory to create/update the man pages.
The script generates the man pages from the `*.adoc` files. Existing files get overridden!
Also the script adds a copyright information at the top of the man page.

After build, mv the man pages are copied to `ospackage/man/`.


# Formatting Rules

- Check that bold is always correct (spaces!) and * as wildcard is not mistakenly used as bold.
- Check if all paths formatted correctly (cursive).
- Check that commands are formatted correctly (bold).
- Check if all variables formatted correctly (cursive).
- Notes and Solution shall be uppercase.
- Make sure the comments `\" _strm_3.2.0_start` and  `\" _strm_3.2.0_end` are preserved. They're needed for package build.
- Make sure the copyright comments are preserved:
  ```
  ++++
  \" _strm_3.2.0_start
  ++++
  ```
  ```
  ++++
  \" _strm_3.2.0_end
  ++++
  ```


