# Quickstart: Testing the CSV Export Feature

This guide provides the steps to test the changes to the client CSV export functionality.

## Prerequisites

- The application must be running locally.
- You must be logged in as a user with permissions to export client data.

## Steps to Test

1.  **Navigate to the Client List**:
    -   Access the client list page in the web application.

2.  **Trigger the Export**:
    -   Click the "Export to CSV" button.
    -   A file named `clients_export.csv` should be downloaded to your computer.

3.  **Verify the CSV Content**:
    -   Open the downloaded `clients_export.csv` file using a spreadsheet application (like Microsoft Excel, Google Sheets, or LibreOffice Calc).
    -   **Check the header row**: It should contain `name`, `email`, and `phone`.
    -   **Verify the absence of the birthdate column**: Confirm that there is no column for `birthdate`.
    -   **Check the data**: Ensure that the client data is correctly populated in the respective columns.

## Expected Result

The downloaded CSV file contains the client data without the `birthdate` column, as specified in the feature requirements.
