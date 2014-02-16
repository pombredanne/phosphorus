# Phosphorus

## Install

	go install willstclair.com/phosphorus

## Usage

	phosphorus command [arguments]

## Commands


	schema -schemadef 'schemadef.json' -source 'sourcedef.json' -o 'file.schema'

Generates a new schema from the given schema definition.

	index -index 'indexdef.json' -source 'sourcedef.json'

Populates the index.

	server -schema 'file.schema' -index 'indexdef.json'

Runs the match server.

## Definition files

### Source definition

#### Syntax

	{
	  "glob": "string",
	  "id_column": 1,
	  "delimiter": "char",
	  "fields": [
		{
		  "name": "first_name",
		  "column": 2
		},
		{
		  "name": "last_name",
		  "column": 3
		}
	  ]
	}

#### Parameters

<dl>
  <dt>glob</dt>
  <dd>Path to input file(s) with wildcard. (e.g.: <tt>/home/foo/data/records_*.csv</tt>)</dd>

  <dt>id_column</dt>
  <dd>Column number of the record ID. IDs must fit in a 32-bit unsigned int. (Columns are one-indexed.)</dd>

  <dt>delimiter</dt>
  <dd>Character used to separate fields. (e.g.: <tt>,</tt> (comma), <tt>\t</tt> (tab))</dd>

  <dt>fields</dt>
  <dd>List of name-column mappings. Names may not be repeated.</dd>
</dl>

### Schema definition

#### Syntax

	{
	  "hash_count": 2048,
	  "chunk_size": 16,
	  "fields": [
		{
		  "comment": "first name",
		  "attrs": [
			"first_name"
		  ],
		  "transforms": [
			{
			  "function": "substr",
			  "arguments": {
				"begin": 0,
				"end": 3
			  }
			},
			{
			  "function": "upcase"
			}
		  ]
		}
	  ]
	}


#### Parameters

<dl>
  <dt>hash_count</dt>
  <dd>Number of hash functions to generate.</dd>

  <dt>chunk_size</dt>
  <dd>Number of hash functions per chunk.</dd>

  <dt>fields</dt>
  <dd>List of field definitions.</dd>
</dl>

##### Field definition
<dl>
  <dt>comment</dt>
  <dd>Comment field.</dd>

  <dt>attrs</dt>
  <dd>List of strings corresponding to fields from the source record. If more than one is specified, the values for each field are concatenated together.</dd>

  <dt>transforms</dt>
  <dd>List of transformations to apply.</dd>
</dl>
