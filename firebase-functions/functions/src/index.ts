import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';

const serviceAccount = require('../circles-4d081-firebase-adminsdk-rtjsi-51616d71b7.json');

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
    const seed = Math.floor(Math.random() * Math.floor(10000));
    admin.firestore().doc('users/' + user.uid).set({
        name: 'user' + seed,
        bio: 'Just Chilling',
        profilepic: 'https://picsum.photos/seed/' + seed + '/300'
    }).then(result => {
        console.log('New User Created');
        console.log(result);
    }).catch(e => {
        console.log(e);
    });

    admin.firestore().doc('socials/' + user.uid).set({
        facebookID: '',
        instagramID: '',
        linkedinID: '',
        numberID: '',
        personalEmailID: '',
        snapchatID: '',
        tiktokID: '',
        twitterID: '',
        venmoID: '',
        websiteID: '',
    }).then(result => {
        console.log('Socials Created');
        console.log(result);
    }).catch(e => {
        console.log(e);
    });

    admin.firestore().doc('preferences/' + user.uid).set({
        preferences: [],
        valueTraits: [],
    }).then(r => {
        console.log('Preferences Created');
        console.log(r);
    }).catch(e => {
        console.log(e);
    });
});

export const sendPing = functions.firestore.document('events/{eventId}').onCreate(async (change, context) => {
    for (const member of change.get('members')) {
        await admin.firestore().collection('pings').add({
            userSent: change.get('creator'),
            userRec: member.ref,
            sentMessage: '',
            responseMessage: 'You\'ve been invited to ' + change.get('name'),
            timeStamp: admin.firestore.FieldValue.serverTimestamp()
        }).catch(e => console.log(e));
    }
});

export const massMessage = functions.https.onCall((data, context) => {
    admin.firestore().collection('events').doc(data.eventId).collection('attendeesPublic')
        .get().then(async r => {
        for (const doc of r.docs) {
            if(<string>context.auth?.uid === doc.id) {
                continue;
            }
            await admin.firestore().collection('pings').add({
                userSent: admin.firestore().collection('users').doc(<string>context.auth?.uid),
                userRec: admin.firestore().collection('users').doc(doc.id),
                sentMessage: '',
                responseMessage: data.message,
                timeStamp: admin.firestore.FieldValue.serverTimestamp()
            }).catch(e => {
                console.log(e);
                return Promise.reject(e);
            });
        }
    }).catch(er => {
        console.log(er);
        return Promise.reject(er);
    });

    return Promise.resolve();
});

