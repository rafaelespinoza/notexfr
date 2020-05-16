```
                  __               ____
   ____   ____ __/ /_ ___   _  __ / __/_____
  / __ \ / __ \\  __// _ \ | |/_// /_ / ___/
 / / / // /_/ // /_ /  __/_>  < / __// /
/_/ /_/ \____/ \__/ \___//_/|_|/_/  /_/
```

`notexfr` is a tool to convert and adapt data for transfer between note-taking services.

#### What are the features?

- Convert Evernote data into StandardNotes format.
- Fetch Note, Notebook, Tag data from your Evernote account (using the EDAM API)
  and write to local JSON files.
- Backfill existing StandardNotes notes with Evernote Notebook metadata.
- Inspect ENEX file (Evernote's export format).

#### Why would you use it?

Use the `convert` tool to transform Evernote data into _new_ StandardNotes data.

If you've already moved your data from Evernote into StandardNotes, you can use
`backfill` to _update_ StandardNotes data. You've probably used the interface
at https://dashboard.standardnotes.org/tools to take an ENEX (Evernote export)
file to convert it to StandardNotes format. Unfortunately, the ENEX format does
not contain any Notebook info, so it can't be preserved with the existing
conversion tool. You'd use this if you want to preserve your Evernote Notebooks
and their Note associations.

## Getting Started

```
go get github.com/rafaelespinoza/notexfr
```

**TLDR**:

- Set up Evernote credentials
- Fetch Evernote data, write to local files
- Convert or backfill StandardNotes data. Do either of the following:
  - Convert data to StandardNotes format (create new data)
  - Backfill data for StandardNotes (update copies of existing data)

### Set up Evernote credentials

You would need a developer token with full access to your account. Visit the
[Evernote developer documentation](https://dev.evernote.com/doc) site.

Request a *developer token* for access to your Evernote account. Typically, you
start with access to a sandbox account, which is for testing things out, and
then request access to your production account. At the time of this writing,
it's a manual process but they are usually pretty quick.

Once you have that info, store them in an environment variable file. Create a
template file and fill it in.

```
notexfr edam make-env
```

### Fetch Evernote data, write to local files

By default, everything is fetched using your sandbox account. Use the
`--production` flag to fetch data from your production Evernote account.

```sh
notexfr edam notebooks --production --output path/to/en_notebooks.json
notexfr edam tags --production --output path/to/en_tags.json
```

To get notes, you should probably have a longer timeout than the default.
Anecdotally, it took about 90 seconds to download about 1550 notes. Your results
will vary. To be safe, set it on the higher end. Add the `verbose` flag
for updates.

```sh
notexfr edam notes \
  --production \
  --timeout 120s \
  --verbose \
  --output path/to/en_notes.json
```

### Convert or backfill StandardNotes data

After downloading your Evernote data to local JSON files, you're ready to
transform it.

##### Convert data to StandardNotes format

_Do this if you want to do a full conversion of Evernote data and create new
StandardNotes data_.

```sh
notexfr convert edam-to-sn \
  --input-en-notebooks path/to/en_notebooks.json \
  --input-en-notes path/to/en_notes.json \
  --input-en-tags path/to/en_tags.json \
  --output path/to/sn.json
```

##### Backfill data for StandardNotes

_Do this if you want to do update existing StandardNotes data_.

```sh
notexfr backfill en-to-sn \
  --input-en-notebooks path/to/en_notebooks.json \
  --input-en-notes path/to/en_notes.json \
  --input-en-tags path/to/en_tags.json \
  --output-notebooks path/to/sn_notebooks.json \
  --output-notes path/to/sn_notes.json \
  --output-tags path/to/sn_tags.json
```
