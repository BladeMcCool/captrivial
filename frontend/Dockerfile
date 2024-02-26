# Start from the node base image
FROM node:latest

# Set the working directory in the Docker container
WORKDIR /app

# Copy the package.json and package-lock.json (if available)
COPY package*.json ./

# Install dependencies
RUN npm install
# Copy the rest of your app's source code
COPY . .

# Expose port 3000 to the outside world
EXPOSE 3000

# Start the app
CMD ["npm", "start"]
