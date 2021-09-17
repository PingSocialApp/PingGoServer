import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';

admin.initializeApp();


// // Start writing Firebase Functions
// // https://firebase.google.com/docs/functions/typescript
//
// export const helloWorld = functions.https.onRequest((request, response) => {
//  response.send("Hello from Firebase!");
// });

export const newUser = functions.runWith({
    memory: '128MB',
    timeoutSeconds: 30
}).auth.user().onCreate((user) => {
    admin.firestore().doc('socials/' + user.uid).set({
        facebook: '',
        instagram: '',
        linkedin: '',
        phone: '',
        personalEmail: '',
        professionalEmail: '',
        snapchat: '',
        tiktok: '',
        twitter: '',
        venmo: '',
        website: '',
    }).then(() =>
        functions.logger.log('Socials Created')
    ).catch((e: any) =>
        functions.logger.error(e)
    );
});

export const handlePings = functions.runWith({
    memory: '512MB',
    timeoutSeconds: 30
}).firestore.document('pings/{docId}').onWrite((change, context) => {
    const databaseRef = admin.database().ref('userNumerics/numPings');

    if (!(change.after.exists)) {     // If deleted ping set userRec to -1
        const previousData = change.before.data();
        if (previousData) {
            return databaseRef.child(previousData.userRec.id).transaction((currentValue) => {
                return (currentValue || 0) <= 0 ? 0 : currentValue - 1
            })
            .then(() => functions.logger.log('Successfully Handled Deleted Ping')).catch(e => functions.logger.error(e));
        }
    } else if (!(change.before.exists)) {     // If new ping set userRec to +1
        const afterData = change.after.data();
        if (afterData) {
            const unreadUpdate = databaseRef.child(afterData.userRec.id)
            .transaction((currentValue) => {
                return (currentValue || 0) < 0 ? 0 : currentValue + 1
            });

            return Promise.all([unreadUpdate, sendNotif(afterData, true)])
            .then(() => functions.logger.log('Successfully Handled Created Ping')).catch(e => functions.logger.error(e));
        }
    } else if (change.after.exists && change.before.exists
        && !change.before.isEqual(change.after)) {     // If updated ping set userSent to -1 and userRec to +1
        const afterData = change.after.data();
        if (afterData) {
            const userRecUpdate = databaseRef.child(afterData.userRec.id)
                .transaction((currentValue) => {
                    return (currentValue || 0) < 0 ? 0 : currentValue + 1
                });
            const userSentUpdate = databaseRef.child(afterData.userSent.id)
                .transaction((currentValue) => {
                    return (currentValue || 0) <= 0 ? 0 : currentValue - 1
                });
            return Promise.all([userRecUpdate, userSentUpdate, sendNotif(afterData, false)])
                .then(() => functions.logger.log('Successfully Handled Replied Ping')).catch(e => functions.logger.error(e));
        }
    }
    return Promise.resolve();
});

function sendNotif(data: any, isNew: boolean): Promise<any> {
    const messaging = admin.messaging();

    return admin.database().ref('notifToken/' + data.userRec.id).once('value', (snapshot) => {
        const token = snapshot.val();
        if (!token) {
            // No token available
            return;
        }
        const payload: admin.messaging.Message = {
            notification: {
                title: isNew ? 'New Ping!' : 'Ping Reply!',
                body: isNew ? `${data.userSent.name} has sent a new ping! 👋` : `${data.userSent.name} has sent a reply 💬`,
                imageUrl: data.userSent.profilepic,
            },
            apns: {
                payload: {
                    aps: {
                        sound: 'default'
                    }
                }
            },
            token
        };
        messaging.send(payload).then((response) =>
            functions.logger.log('Successfully set message:', response)
        ).catch(err => functions.logger.error(err));

    }, (err) => {
        functions.logger.error(err);
    });
}





