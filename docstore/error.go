package docstore

type DocstoreError string

func (e DocstoreError) Error() string { return string(e) }

const NotFound = DocstoreError("[docstore] document not found")
