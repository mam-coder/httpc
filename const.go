package httpc

// Common Content-Type header values for HTTP requests and responses.
const (
	// ContentTypeJSON represents JSON content (application/json)
	ContentTypeJSON = "application/json"

	// ContentTypeXML represents XML content (application/xml)
	ContentTypeXML = "application/xml"

	// ContentTypeForm represents URL-encoded form data (application/x-www-form-urlencoded)
	ContentTypeForm = "application/x-www-form-urlencoded"

	// ContentTypeMultipart represents multipart form data (multipart/form-data)
	ContentTypeMultipart = "multipart/form-data"

	// ContentTypePlainText represents plain text content (text/plain)
	ContentTypePlainText = "text/plain"

	// ContentTypeHTML represents HTML content (text/html)
	ContentTypeHTML = "text/html"

	// ContentTypeCSV represents CSV content (text/csv)
	ContentTypeCSV = "text/csv"

	// ContentTypeApplicationCSV represents CSV content (application/csv)
	ContentTypeApplicationCSV = "application/csv"

	// ContentTypeJavaScript represents JavaScript content (application/javascript)
	ContentTypeJavaScript = "application/javascript"

	// ContentTypeCSS represents CSS content (text/css)
	ContentTypeCSS = "text/css"

	// ContentTypePDF represents PDF content (application/pdf)
	ContentTypePDF = "application/pdf"

	// ContentTypeZip represents ZIP archive content (application/zip)
	ContentTypeZip = "application/zip"

	// ContentTypeOctetStream represents arbitrary binary data (application/octet-stream)
	ContentTypeOctetStream = "application/octet-stream"
)
