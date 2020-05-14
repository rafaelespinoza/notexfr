```
                  __               ____
   ____   ____ __/ /_ ___   _  __ / __/_____
  / __ \ / __ \\  __// _ \ | |/_// /_ / ___/
 / / / // /_/ // /_ /  __/_>  < / __// /
/_/ /_/ \____/ \__/ \___//_/|_|/_/  /_/
```

`notexfr` is a tool to derive untransferred metadata when exporting Evernote
data into [StandardNotes](https://standardnotes.org/), such as Notebooks and
their associated Notes. It's not meant to replace the [existing data
tools](https://dashboard.standardnotes.org/tools). Rather it fills in some gaps.
Think of it as a backfill.

#### What are the features?

- Format Evernote notebooks into data suitable for import into StandardNotes.
- Re-associate StandardNotes notes with Evernote Notebook metadata.
- Identify Evernote tags, notes that correspond to StandardNotes tags, notes.
- Fetch Note, Notebook, Tag data from your Evernote account (using the EDAM API)
  and write to local JSON files.
- Inspect ENEX file (Evernote's export format).

#### Why would you use it?

You'd use this if you want to preserve your Evernote Notebooks and their Note
associations after converting them to the StandardNotes format

#### Why do you need it?

The conversion tool uses an ENEX (evernote export) file to convert it to
StandardNotes format. Unfortunately, the ENEX format does not contain any
Notebook info, so it can't be preserved with the existing conversion tool.

#### How would you use it?

It's meant to be used locally. For example, you use your own Evernote API key to
download Notebook, Note metadata and save to local files. Take the Evernote
metadata to lookup Note to Notebook associations and then append references to
StandardNotes data.

## Getting Started

```
go get github.com/rafaelespinoza/notexfr
```

#### Evernote API

You would need a developer token with full access to your account. This is so
you can download the data necessary to make associations to your standardnotes
account data. Part of that includes reading existing content, which requires
full access to your account.

Visit the [Evernote developer documentation](https://dev.evernote.com/doc) site.

Request a *developer token* for access to your Evernote account. Typically, you
start with access to a sandbox account, which is for testing things out, and
then request access to your production account. At the time of this writing,
it's a manual process but they are usually pretty quick.

Once you have that info, start an environment file template.

```
notexfr edam make-env
```

This creates a file at `.env`. Fill it in.
