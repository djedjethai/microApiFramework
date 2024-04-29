FROM nginx:1.24 

WORKDIR /

# COPY ./default.conf /etc/nginx/conf.d/default.conf
# COPY ./nginx.conf /etc/nginx/nginx.conf

WORKDIR /nginxlogs
COPY ./nginxlogs .

WORKDIR /

CMD ["nginx", "-g", "daemon off;"]

