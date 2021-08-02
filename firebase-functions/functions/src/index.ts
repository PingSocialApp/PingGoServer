import * as functions from 'firebase-functions';
import * as admin from 'firebase-admin';

admin.initializeApp()


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

export const updateNumPings = functions.runWith({
    memory: '128MB',
    timeoutSeconds: 30
}).firestore.document('pings/{docId}').onWrite((change,context) => {

    const databaseRef = admin.database().ref("userNumerics").child('numPings');

    if(!(change.after.exists)){     //If deleted ping set userRec to -1
        const previousData = change.before.data();
        if(previousData){
            databaseRef.child(previousData.userRec).transaction((current_value) => (current_value || 0) <= 0 ? 0 : current_value-1)
            .then(() => functions.logger.log('Successfully Handled Deleted Ping')).catch(e => functions.logger.error(e));
        }
    } else if(!(change.before.exists)){     //If new ping set userRec to +1
        const afterData = change.after.data();
        if(afterData){
            databaseRef.child(afterData.userRec).transaction((current_value) => (current_value || 0) < 0 ? 0 : current_value+1)
            .then(() => functions.logger.log('Successfully Handled Created Ping')).catch(e => functions.logger.error(e));
        }
    } else if(change.after.exists && change.before.exists && !change.before.isEqual(change.after)){     //If updated ping set userSent to -1 and userRec to +1
        const afterData = change.after.data();
        if(afterData){
            const userRecUpdate = databaseRef.child(afterData.userRec).transaction((current_value) => (current_value || 0) < 0 ? 0 : current_value+1);
            const userSentUpdate = databaseRef.child(afterData.userSent).transaction((current_value) => (current_value || 0) <= 0 ? 0 : current_value-1);
            Promise.all([userRecUpdate, userSentUpdate]).then(() => functions.logger.log('Successfully Handled Replied Ping')).catch(e => functions.logger.error(e));
        }
    }
});


