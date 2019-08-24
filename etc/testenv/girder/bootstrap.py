import requests

API_URL = 'http://localhost:8080/api/v1'

user_exists = requests.get(API_URL + '/user/authentication', auth=('admin', 'password'))

if user_exists.ok:
    token = user_exists.json()['authToken']['token']
else:
    user = requests.post(
        API_URL + '/user',
        data={
            'login': 'admin',
            'email': 'girder@girder.girder',
            'firstName': 'girder',
            'lastName': 'girder',
            'password': 'password',
        },
    )
    assert user.ok

    token = user.json()['authToken']['token']

assetstores = requests.get(API_URL + '/assetstore', headers={'Girder-Token': token})
assert assetstores.ok

if not len(assetstores.json()):
    set_assetstore = requests.post(
        API_URL + '/assetstore',
        data={'name': 'assetstore', 'type': 0, 'root': '/assetstore'},
        headers={'Girder-Token': token},
    )
    assert set_assetstore.ok
