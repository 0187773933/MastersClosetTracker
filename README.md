# Master's Closet Tracking Server

## Onboarding Experience
1. Admin Enters Provided First and Last Name
2. Server Redirects to `/admin/user/new/handoff/${new-users-uuid}`
3. New user scans Hand-Off QR code with their phone
4. Scanned QR Hand-Off Code takes them to a silent login page that stores a permanent login cookie.
	- `/user/login/fresh/${new-users-uuid}`

---

## To Re-Enter
1. They scan a QR code on a poster at the front door or just go to `/checkin`
2. If they have a cookie stored it redirects to `/user/checkin/display/${uuid}`
3. Admin Scans and checks-in/validates their QR-Code with stored uuid

---

## TODO

1. Fix "Clothes Balance"
	- calculate based off party total
	- give option on check-in to increase or decrease
	- use printout reciept tickets with qrcode / barcode
	- Checkout counter scans / verifies reciept to match balance
2. Add running daily total screen
3. Just let a barcode check-in a user. Avoids an extra call
	- GET /admin/user/get/barcode/:barcode
	- GET /admin/user/checkin/test/:uuid
	- GET /admin/user/checkin/:uuid
4. Add Admin Manual Override Routes
	- Override Check-In Too Soon
	- User forgot phone
	- User has new phone
	- option to text hand-off link if user can't scan qrcode for some reason
5. Add Spreadsheet Export
6. Fix User Fields :
	- Authorized Aliases
7. Fix Docker
8. Why aren't we using any of the built-in time functions ?
	- `time.Now().After(lastFetched.Add(CachePeriod))` ?
9. Change "usernames" DB bucket for key=${uuid}_username , value=Username
	- keeps only uuids as keys