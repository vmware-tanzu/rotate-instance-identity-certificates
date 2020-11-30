# Docs

This directory contains the docs that are packaged up as a
[hugo](https://gohugo.io/) site.

## Getting Started

- Install Hugo: as simple as `brew install hugo` on a Mac
- If you've just cloned this repo, ensure you've got the theme as well
  by running the following: `git submodule init && git submodule update`

## How Do I

### Add a New Page

To generate the new page, run: `hugo new docs/some-name.md`.
This will place a file under `content/docs`.

Note that the new file may say `draft: true` in the _frontmatter_.
Drafts are not deployed by default, but can be viewed locally by
running `hugo server -D`.

When you're ready to include the new page in the docs, set `draft: false`.

### View the Docs Locally

Start a dev server using `hugo server -D` and point your browser at localhost.
Make sure you're in the `riic-docs` directory.

The dev server will automatically rebuild the site when you save a file,
so feel free to leave it running as you edit.

### View the Live Docs Online

The docs are hosted on Github Pages at https://vmware-tanzu.github.io/rotate-instance-identity-certificates.

### Build the Docs

From the `riic-docs` directory, run `hugo`.
This regenerates the files in the `riic-docs/public` directory.

### Deploy the Docs

The docs are deployed automatically on commit via Github actions.

### Update the Theme?

We're using the [Whisper Theme](https://themes.gohugo.io/hugo-whisper-theme/).
The theme is a git submodule in the `themes/` directory.
To update, simply run `git submodule update --remote --merge`.

### Add a Section to the Nav Bar

The menu is defined in `config.toml`.
To create a new section, add a new `[[menu.main]]` section.
