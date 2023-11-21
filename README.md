# Sanity Translate Service

Sanity Translate Service is a robust document translation tool designed to integrate seamlessly with Sanity.io's content management platform. This Go application serves as a middleware, facilitating the translation of documents stored within Sanity using the DeepL API for language processing.

## Features

- Document Translation: Translates documents from one language to another using DeepL API, while preserving formatting.
- Sanity Integration: Directly interfaces with Sanity's data mutation and query APIs to fetch and update documents.
- CORS Support: Configurable CORS policy for cross-origin resource sharing.
- Input Flexibility: Supports translation of specified document elements or fields.
- Translation Memory: Maintains a translation metadata document within Sanity to keep track of translations.
- API Endpoint: Exposes a POST endpoint for translating documents via HTTP requests.

## Getting Started

### Prerequisites

- Go (version 1.15+ recommended)
- Sanity.io project with an associated project ID, version, and token
- DeepL API account with an authorization token

### Installation

1. Clone the repository to your local machine.
2. Set your Sanity and DeepL API credentials as environment variables:

```bash
export SANITY_PROJECT_ID="your_sanity_project_id"
export SANITY_VERSION="your_sanity_version"
export SANITY_TOKEN="your_sanity_token"
export DEEPL_TOKEN="your_deepl_auth_key"
```

3. Navigate to the project directory and build the application:

```bash
go build
```

4. Run the application:

```bash
./SanityTranslator
```

## Usage

Send a POST request to the /sanity_translate_document endpoint with the JSON payload specifying the document slugs and the target language. 
For example:

```json
{
    "FromSlug": "/it/this-is-the-slug-from",
    "ToSlug": "/de/this-is-the-slug-to",
    "Lang": "de",
    "InputElements": [
        "title",
        "subtitle",
        "intro",
        "metadata.metaKeyword",
        "metadata.metaDescription",
        "metadata.metaTitle",
        "text.000.title",
        "text.000.intro ",
        "text.000.text",
        "text.000.children.000.text",
    ]
}
```

Here how a working CURL request look like:

```bash
curl --location 'localhost:8080/sanity_translate_document' \
--header 'Content-Type: application/json' \
--data '{
    "FromSlug": "/it",
    "ToSlug": "/de",
    "Lang": "de",
    "InputElements": [
        "title",
        "intro",
        "metadata.metaKeyword",
        "metadata.metaDescription",
        "metadata.metaTitle",
        "text.000.title",
        "text.000.intro ",
        "text.000.children.000.text"
    ]
}'
```

The service will fetch the specified document from Sanity, translate the designated elements, and create a new translated document in the target language.

## Development

To contribute to the development of Sanity Translate Service, you can:

- Fork the repository.
- Create a new feature branch.
- Commit your changes.
- Submit a pull request.

## Compatibility and Considerations

### Sanity V3 and Internationalization

SanityTranslator is designed to operate seamlessly with Sanity V3 and is fully compatible with the latest version of the Sanity.io Internationalization Plugin. This ensures that the tool leverages the most up-to-date features for content translation and internationalization offered by Sanity.

### Document Overwrite Warning

Please be aware that using SanityTranslator to translate a document will overwrite any previously created document with the same identifier. Exercise caution and ensure that you have appropriate backups or versioning in place before you translate and overwrite existing content.

### Testing

Currently, SanityTranslator lacks a suite of automated tests. As such, it is recommended that users manually verify the tool’s functionality in a controlled environment before deploying it in a production setting. Contributions to develop a comprehensive testing suite are welcome and encouraged.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
