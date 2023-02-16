# Master's Closet Tracking Server

## Onboarding Experience
1. Admin Enters Provided First and Last Name
2. Server Redirects to `/admin/user/new/handoff/${new-users-uuid}`
3. New user scans Hand-Off QR code with their phone
4. Scanned QR Hand-Off Code takes them to a silent login page that stores a permanent login cookie.
	- `/user/login/${new-users-uuid}`

---

## To Re-Enter
1. They scan a QR code on a poster at the front door or just go to `/checkin`
2. If they have a cookie stored it redirects to `/user/checkin/display/${uuid}`
3. Admin Scans and checks-in/validates their QR-Code with stored uuid

---

## TODO

1. Admin Routes for Manual Overrides
	- User forgot phone
	- User has new phone
	- option to text hand-off link if user can't scan qrcode for some reason
2. Add and Fix Time to Re-Checkin
3. Add Spreadsheet Export
4. Fix Edge Cases
	- User attempts to use a different name ? FaceID ?
5. Fix Docker