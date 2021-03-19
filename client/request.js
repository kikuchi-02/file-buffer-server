import axios from 'axios';

const req = (k, i=0) =>{
        return axios.create({
            headers: {
                'HTTP_CLOUDFRONT_VIEWER_COUNTRY': 'country',
                'X-aaaaa': 'test'
            }
        }).post('http://localhost:8000/eventlog', 
        {
            request_method: 'request method',
            user_agent: `agent-${k}-${i}`,
            logs: [{
                created:  Date.now(),
            }*100]
        }
        )
}

(async () => {
    const res = await req(1)
    console.log(res.data)
    for (let i = 0; i<10000000;i++){
        if (i%100===0){
        console.log('request', i)

        }
        await Promise.all(Array.from(Array(5), (_v, _k) => _k).map((_k) =>{
            return req(i, _k)
        }))
    }

})()