# Go Wiki

## Vision

Create a wiki for storing requirements, design documents, test cases (executable cucumber?)
with tracability between these.

## TODO
 * Make it look good

 * Webs
 * Page Meta Data (_parent)
 * WebPreferences / SitePreferences
 * User Management
 * Groups and Permissions
 * Revisions
 * Backlinks
 * Attachments
 * Print Skin
 * Authentication with Google
 * Forms
 * Anchors on headings
 * Plugins (TOC, Search)
 * Code Syntax Highlighting


## User Registration
 * on initial set up, not authenticated (Main.WikiPreferences  * set RequireAuthentication=off)
 
To allow registration without requesting access...
 * to enable a page to be shown non authenticated add * set RequireAuthentication=off (can be done on UserRegistration)

On Registration... create a UserPage and add to WikiUsers
On creation has a Parent page (set to WikiUsers) (Use MetaData)

Where to store password? (use MetaData on UserPage... needs to be encrypted)


## Reset Password...? TOTP from email link
