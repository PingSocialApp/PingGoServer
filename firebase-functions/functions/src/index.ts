import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';

const serviceAccount = require('../circles-4d081-firebase-adminsdk-rtjsi-6ab3240fd0.json');

admin.initializeApp({
    credential: admin.credential.cert(serviceAccount),
    databaseURL: `https://${serviceAccount.project_id}.firebaseio.com`
});


// // Start writing Firebase Functions
// // https://firebase.google.com/docs/functions/typescript
//
// export const helloWorld = functions.https.onRequest((request, response) => {
//  response.send("Hello from Firebase!");
// });

export const newUser = functions.auth.user().onCreate((user) => {
    admin.firestore().doc('socials/' + user.uid).set({
        facebook: '',
        instagram: '',
        linkedin: '',
        phone: '',
        personalEmail: '',
        snapchat: '',
        tiktok: '',
        twitter: '',
        venmo: '',
        website: '',
    }).then((result: any) => {
        console.log('Socials Created');
        console.log(result);
    }).catch((e: any) => {
        console.log(e);
    });
});


